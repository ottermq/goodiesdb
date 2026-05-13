package command

import (
	"strings"

	"github.com/ottermq/goodiesdb/internal/protocol"
)

type EchoCommand struct{}

func NewEchoCommand() *EchoCommand {
	return &EchoCommand{}
}

func (c *EchoCommand) Name() string {
	return "ECHO"
}

func (c *EchoCommand) RequiresAuth() bool {
	return false
}

func (c *EchoCommand) Validate(args []string) error {
	if len(args) < 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (c *EchoCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	return protocol.BulkString([]byte(strings.Join(args, " "))), nil
}
