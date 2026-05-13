package command

import (
	"fmt"

	"github.com/ottermq/goodiesdb/internal/protocol"
)

type SelectCommand struct{}

func NewSelectCommand() *SelectCommand {
	return &SelectCommand{}
}

func (c *SelectCommand) Name() string {
	return "SELECT"
}

func (c *SelectCommand) RequiresAuth() bool {
	return false
}

func (c *SelectCommand) Validate(args []string) error {
	return requireExactArgs(args, 1)
}

func (c *SelectCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	dbIndex, err := parseIntArg(args[0], "invalid DB index")
	if err != nil {
		return nil, err
	}
	if ctx.SelectDB == nil {
		return nil, fmt.Errorf("select DB not available")
	}
	if err := ctx.SelectDB(dbIndex); err != nil {
		return nil, err
	}
	return protocol.SimpleString("OK"), nil
}
