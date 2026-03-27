package command

import (
	"fmt"
	"strconv"

	"github.com/andrelcunha/goodiesdb/internal/protocol"
)

type SelectCommand struct{}

func NewSelectCommand() *SelectCommand {
	return &SelectCommand{}
}

func (c *SelectCommand) Name() string {
	return "SELECT"
}

func (c *SelectCommand) MinArgs() int {
	return 1
}

func (c *SelectCommand) MaxArgs() int {
	return 1
}

func (c *SelectCommand) RequiresAuth() bool {
	return false
}

func (c *SelectCommand) Validate(args []string) error {
	if len(args) != 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (c *SelectCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	dbIndex, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, fmt.Errorf("invalid DB index")
	}
	if ctx.SelectDB == nil {
		return nil, fmt.Errorf("select DB not available")
	}
	if err := ctx.SelectDB(dbIndex); err != nil {
		return nil, err
	}
	return protocol.SimpleString("OK"), nil
}
