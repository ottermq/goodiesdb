package server

import (
	"fmt"
	"net"
	"path/filepath"
	"time"

	"github.com/ottermq/goodiesdb/internal/logging"
	"github.com/ottermq/goodiesdb/internal/persistence/aof"
	"github.com/ottermq/goodiesdb/internal/persistence/rdb"
)

func (s *Server) isAuthenticates(conn net.Conn) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.connections[conn]
	return ok && c.authed
}

func (s *Server) Authenticate(conn net.Conn, password string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if password != s.config.Password {
		return false
	}
	if c, ok := s.connections[conn]; ok {
		c.authed = true
	}
	return true
}

func (s *Server) getCurrentDb(conn net.Conn) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if c, ok := s.connections[conn]; ok {
		return c.dbIndex
	}
	return 0
}

// Quit closes the connection
func (s *Server) Quit(conn net.Conn) {
	fmt.Fprintln(conn, "OK")
	conn.Close()
}

// SelectDb selects the database
func (s *Server) SelectDb(conn net.Conn, dbIndex int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if dbIndex < 0 || dbIndex >= s.store.Count() {
		return fmt.Errorf("invalid DB index")
	}
	if c, ok := s.connections[conn]; ok {
		c.dbIndex = dbIndex
	}
	return nil
}

func (s *Server) Addr() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener == nil {
		return ""
	}

	return s.listener.Addr().String()
}

func (s *Server) WaitForReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		addr := s.Addr()
		if addr != "" {
			conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
			if err == nil {
				_ = conn.Close()
				return nil
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("server did not become ready within %s", timeout)
}

func (s *Server) stopAcceptLoop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.shutdownChan:
	default:
		close(s.shutdownChan)
	}

	if s.listener != nil {
		_ = s.listener.Close()
	}
}

func (s *Server) startRDB() {
	rdbFilepath := filepath.Join(s.dataDir, "dump.rdb")
	for {
		select {
		case <-time.After(1 * time.Minute):
			if err := rdb.SaveSnapshot(s.store, rdbFilepath); err != nil {
				logging.Errorf("Error saving snapshot: %v", err)
			} else {
				logging.Infof("Snapshot saved successfully")
			}

		case <-s.shutdownChan:
			return
		}
	}
}

func (s *Server) recoverStore() {
	rdbFilepath := filepath.Join(s.dataDir, "dump.rdb")
	aofFilepath := filepath.Join(s.dataDir, "appendonly.aof")
	flagOk := false
	if s.config.UseRDB {
		if err := rdb.LoadSnapshot(s.store, rdbFilepath); err != nil {
			logging.Infof("No snapshot found.")
		} else {
			flagOk = true
		}
	}

	if s.config.UseAOF && !flagOk {
		if err := aof.RebuildStoreFromAOF(s.store, aofFilepath); err != nil {
			logging.Errorf("Error loading from AOF: %v", err)

		} else {
			flagOk = true
		}
	}
	if !flagOk {
		logging.Infof("None of the recovery files are healthy. Starting with an empty store.")
	}
}

func (s *Server) asciiLogo() string {
	return `
 ██████╗  ██████╗  ██████╗ ██████╗ ██╗███████╗███████╗██████╗ ██████╗
██╔════╝ ██╔═══██╗██╔═══██╗██╔══██╗██║██╔════╝██╔════╝██╔══██╗██╔══██╗
██║  ███╗██║   ██║██║   ██║██║  ██║██║█████╗  ███████╗██║  ██║██████╔╝
██║   ██║██║   ██║██║   ██║██║  ██║██║██╔══╝  ╚════██╗██║  ██║██╔══██╗
╚██████╔╝╚██████╔╝╚██████╔╝██████╔╝██║███████╗███████║██████╔╝██████╔╝
 ╚═════╝  ╚═════╝  ╚═════╝ ╚═════╝ ╚═╝╚══════╝╚══════╝╚═════╝ ╚═════╝
        RESP-compatible · in-memory · blazing fast
`
}
