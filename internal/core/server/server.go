package server

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/andrelcunha/goodiesdb/internal/core/command"
	"github.com/andrelcunha/goodiesdb/internal/core/store"
	"github.com/andrelcunha/goodiesdb/internal/logging"
	"github.com/andrelcunha/goodiesdb/internal/persistence/aof"
	"github.com/andrelcunha/goodiesdb/internal/persistence/rdb"
	"github.com/andrelcunha/goodiesdb/internal/protocol"
	"github.com/andrelcunha/goodiesdb/internal/protocol/resp2"
)

// Server represents a TCP server
type Server struct {
	store           *store.Store
	config          *Config
	mu              sync.Mutex
	connections     map[net.Conn]*Conn
	broker          *PubSubBroker
	shutdownChan    chan struct{}
	dataDir         string
	Protocol        protocol.Protocol
	commandRegistry *command.Registry
	listener        net.Listener
	connCounter     int64
}

// NewServer creates a new server
func NewServer(config *Config) *Server {
	// Create the data directory if it doesn't exist
	dataDir := config.DataDir
	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		fmt.Printf("Error creating data directory: %v\n", err)
		os.Exit(1)
	}
	protocol := &resp2.RESP2Protocol{}

	aofChan := make(chan store.AOFCommand, 100)
	s := store.NewStore(aofChan)
	server := &Server{
		store:           s,
		config:          config,
		connections:     make(map[net.Conn]*Conn),
		broker:          newPubSubBroker(),
		shutdownChan:    make(chan struct{}),
		dataDir:         config.DataDir,
		Protocol:        protocol,
		commandRegistry: command.NewRegistry(),
	}
	return server
}

// Start starts the server
func (s *Server) Start() error {
	logging.Infof("%s", s.asciiLogo())
	logging.Infof("Starting GoodiesDB server...")

	if s.config.UseRDB || s.config.UseAOF {
		logging.Infof("Found persistence enabled. Recovering data...")
		s.recoverStore()
	} else {
		logging.Infof("No persistence enabled. Data will not be persisted.")
	}

	if s.config.UseRDB {
		go s.startRDB()
		logging.Infof("RDB persistence enabled")
	}
	if s.config.UseAOF {
		aofFilepath := filepath.Join(s.dataDir, "appendonly.aof")
		go aof.AOFWriter(s.store.AOFChannel(), aofFilepath)
		logging.Infof("AOF persistence enabled")
	}

	// set addr string (host and port) using config
	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.listener = ln
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.listener = nil
		s.mu.Unlock()
		_ = ln.Close()
	}()
	logging.Infof("GoodiesDB %s listening on %s", s.config.Version, ln.Addr().String())

	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			logging.Errorf("Error accepting connection: %v", err)
			continue
		}
		go s.handleConn(conn)
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() {
	s.stopAcceptLoop()

	if s.config.UseAOF {
		if s.store.AOFChannel() != nil {
			close(s.store.AOFChannel())
		}
	}

	if s.config.UseRDB {
		rdb.SaveSnapshot(s.store, "dump.rdb")
	}
}

// allowedInSubMode is the set of commands permitted while a connection is in
// subscriber mode. All others are rejected with an error.
var allowedInSubMode = map[string]bool{
	"SUBSCRIBE":    true,
	"UNSUBSCRIBE":  true,
	"PSUBSCRIBE":   true,
	"PUNSUBSCRIBE": true,
	"PING":         true,
	"QUIT":         true,
}

func (s *Server) handleConn(conn net.Conn) {
	s.mu.Lock()
	s.connCounter++
	c := newConn(conn)
	c.id = s.connCounter
	s.connections[conn] = c

	s.mu.Unlock()

	defer func() {
		// UnsubscribeAll closes the broker delivery channel, which terminates
		// the write goroutine (if one was started) via range-channel exhaustion.
		s.broker.UnsubscribeAll(conn)
		s.mu.Lock()
		delete(s.connections, conn)
		s.mu.Unlock()
		conn.Close()
	}()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		value, err := s.Protocol.Parse(reader)
		if err != nil {
			if err.Error() == "EOF" {
				return
			}
			reply := protocol.ErrorString(fmt.Sprintf("parse error: %v", err))
			s.Protocol.Encode(writer, reply)
			writer.Flush()
			continue
		}

		// Check subscriber mode restrictions before executing.
		if c.mode == ModeSubscriber {
			cmdName := ""
			if arr, ok := value.(protocol.Array); ok && len(arr) > 0 {
				if bs, ok := arr[0].(protocol.BulkString); ok {
					cmdName = strings.ToUpper(string(bs))
				}
			}
			if !allowedInSubMode[cmdName] {
				s.Protocol.Encode(writer, protocol.ErrorString("ERR Command not allowed in subscribe mode"))
				writer.Flush()
				continue
			}
			// PING in subscriber mode returns a push array, not +PONG.
			if cmdName == "PING" {
				writer.Write([]byte("*3\r\n$4\r\npong\r\n$0\r\n\r\n"))
				writer.Flush()
				continue
			}
		}

		reply, err := s.executeCommand(conn, value)
		if err != nil {
			reply := protocol.ErrorString(fmt.Sprintf("ERR %s", err.Error()))
			s.Protocol.Encode(writer, reply)
			writer.Flush()
			continue
		}

		if reply != nil {
			s.Protocol.Encode(writer, reply)
			writer.Flush()
		}
	}
}

func (s *Server) executeCommand(conn net.Conn, request protocol.RESPValue) (protocol.RESPValue, error) {
	arr, ok := request.(protocol.Array)
	if !ok {
		return protocol.ErrorString("ERR expected array"), fmt.Errorf("expected array, got %T", request)
	}
	parts := convertArrayToStrings(arr)

	if len(parts) == 0 {
		return protocol.SimpleString(""), nil
	}

	cmdName := strings.ToUpper(parts[0])
	args := parts[1:]
	logging.Debugf("Executing command: %s %v", cmdName, args)
	dbIndex := s.getCurrentDb(conn)
	/* Using registered commands*/

	cmd, ok := s.commandRegistry.Get(cmdName)
	if ok {
		return s.invokeCommand(cmd, args, conn, dbIndex)
	}
	return protocol.ErrorString("ERR unknown command '" + parts[0] + "'"), nil
}

func (s *Server) invokeCommand(cmd command.Command, args []string, conn net.Conn, dbIndex int) (protocol.RESPValue, error) {
	if err := cmd.Validate(args); err != nil {
		return protocol.ErrorString("ERR " + err.Error()), nil
	}

	// Check authentication
	if cmd.RequiresAuth() && !s.isAuthenticates(conn) {
		return protocol.ErrorString("NOAUTH Authentication required"), nil
	}

	s.mu.Lock()
	c := s.connections[conn]
	s.mu.Unlock()

	ctx := &command.Context{
		Store:     s.store,
		DBIndex:   dbIndex,
		Conn:      conn,
		Timestamp: time.Now(),
		Nil: func() protocol.RESPValue {
			return s.Protocol.EncodeNil()
		},
		Auth: func(password string) bool {
			return s.Authenticate(conn, password)
		},
		SelectDB: func(index int) error {
			return s.SelectDb(conn, index)
		},
		Info: func() protocol.BulkString {
			return s.Info()
		},
		// Region Client _commands
		GetConnID: func() int64 {
			s.mu.Lock()
			defer s.mu.Unlock()
			if c == nil {
				return 0
			}
			return c.id
		},
		GetConnName: func() string {
			s.mu.Lock()
			defer s.mu.Unlock()
			if c == nil {
				return ""
			}
			return c.name
		},
		SetConnName: func(name string) {
			s.mu.Lock()
			defer s.mu.Unlock()
			if c == nil {
				return
			}
			c.name = name
		},
		GetConnInfo: func() string {
			s.mu.Lock()
			defer s.mu.Unlock()
			if c == nil {
				return ""
			}
			return fmt.Sprintf("id=%d addr=%s name=%s db=%d", c.id, conn.RemoteAddr(), c.name, c.dbIndex)
		},
		// EndRegion
		PubSub: s.broker,
		Write: func(v protocol.RESPValue) {
			// Safe to write directly here: the write goroutine is only started
			// after SetMode(ModeSubscriber) returns, so no concurrent writer yet.
			w := bufio.NewWriter(conn)
			s.Protocol.Encode(w, v)
			w.Flush()
		},
		SetMode: func(mode int) {
			s.mu.Lock()
			defer s.mu.Unlock()
			if c == nil {
				return
			}
			newMode := ConnMode(mode)
			if newMode == ModeSubscriber && c.mode != ModeSubscriber {
				// First transition: get the delivery channel the broker created
				// during Subscribe/PSubscribe and start the write goroutine.
				deliveryCh := s.broker.GetConnChan(conn)
				if deliveryCh != nil {
					c.mode = ModeSubscriber
					go func() {
						for msg := range deliveryCh {
							conn.Write(msg)
						}
					}()
				}
			} else if newMode == ModeNormal {
				c.mode = ModeNormal
			}
		},
	}

	return cmd.Execute(ctx, args)
}

func convertArrayToStrings(rawParts protocol.Array) []string {
	parts := make([]string, len(rawParts))
	for i, part := range rawParts {
		switch v := part.(type) {
		case protocol.BulkString:
			parts[i] = string(v)
		case protocol.SimpleString:
			parts[i] = string(v)
		case string:
			parts[i] = v
		default:
			// Fallback: convert to string
			parts[i] = fmt.Sprintf("%v", v)
		}
	}
	return parts
}
