package command

import (
	"fmt"
	"net"
	"time"

	"github.com/andrelcunha/goodiesdb/internal/core/store"
	"github.com/andrelcunha/goodiesdb/internal/protocol"
)

type Command interface {
	Name() string
	Execute(ctx *Context, args []string) (protocol.RESPValue, error)
	Validate(args []string) error
	RequiresAuth() bool
}

// PubSubBroker is the interface pub/sub commands use to interact with the broker.
// Defined here to avoid an import cycle between command and server packages.
type PubSubBroker interface {
	Subscribe(conn net.Conn, channel string) int
	Unsubscribe(conn net.Conn, channel string) int
	PSubscribe(conn net.Conn, pattern string) int
	PUnsubscribe(conn net.Conn, pattern string) int
	Publish(channel, message string) int
	UnsubscribeAll(conn net.Conn)
}

type Context struct {
	Store     *store.Store
	DBIndex   int
	Conn      net.Conn
	Timestamp time.Time
	Nil       func() protocol.RESPValue
	Auth      func(password string) bool
	SelectDB  func(dbIndex int) error
	Info      func() protocol.BulkString
	Protocol  protocol.Protocol
	// Client command
	GetConnID   func() int64
	GetConnName func() string
	SetConnName func(name string)
	GetConnInfo func() string
	// Pub/sub support
	PubSub  PubSubBroker
	Write   func(v protocol.RESPValue) // write a response directly, bypassing normal return
	SetMode func(mode int)             // flip connection mode (0 = normal, 1 = subscriber)
}

var ErrWrongNumberOfArguments = fmt.Errorf("ERR wrong number of arguments for command")
