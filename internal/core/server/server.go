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
	store                    *store.Store
	config                   *Config
	mu                       sync.Mutex
	authenticatedConnections map[net.Conn]bool // TODO create a connection abstraction to hold more info
	connectionDbs            map[net.Conn]int
	shutdownChan             chan struct{}
	dataDir                  string
	Protocol                 protocol.Protocol
	commandRegistry          *command.Registry
	listener                 net.Listener
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
		store:                    s,
		config:                   config,
		authenticatedConnections: make(map[net.Conn]bool),
		connectionDbs:            make(map[net.Conn]int),
		shutdownChan:             make(chan struct{}),
		dataDir:                  config.DataDir,
		Protocol:                 protocol,
		commandRegistry:          command.NewRegistry(),
	}
	return server
}

// Start starts the server
func (s *Server) Start() error {
	logging.Infof("%s", s.asciiLogo())
	logging.Infof("Starting Redis Clone Server...")

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
	logging.Infof("Redis Clone Server %s started on %s", s.config.Version, ln.Addr().String())

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

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
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

		// Execute commmand
		reply, err := s.executeCommand(conn, value)
		if err != nil {
			reply := protocol.ErrorString(fmt.Sprintf("ERR %s", err.Error()))
			s.Protocol.Encode(writer, reply)
			writer.Flush()
			continue
		}

		s.Protocol.Encode(writer, reply)
		writer.Flush()
		continue
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
