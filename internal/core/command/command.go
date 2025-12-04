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
	MinArgs() int
	MaxArgs() int
}

type Context struct {
	Store     *store.Store
	DBIndex   int
	Conn      net.Conn
	Timestamp time.Time
}

var ErrWrongNumberOfArguments = fmt.Errorf("ERR wrong number of arguments for command")
