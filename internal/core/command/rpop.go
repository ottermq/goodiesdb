package command

import (
	"fmt"
	"strconv"

	"github.com/andrelcunha/goodiesdb/internal/protocol"
)

type RPopCommand struct{}

func NewRPopCommand() *RPopCommand {
	return &RPopCommand{}
}

func (c *RPopCommand) Name() string {
	return "RPOP"
}

func (c *RPopCommand) MinArgs() int {
	return 1
}

func (c *RPopCommand) MaxArgs() int {
	return 2
}

func (c *RPopCommand) RequiresAuth() bool {
	return false // TODO: refactor authentication
}

func (c *RPopCommand) Validate(args []string) error {
	if len(args) != 1 && len(args) != 2 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (c *RPopCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	var count *int
	if len(args) == 2 {
		parsed, err := strconv.Atoi(args[1])
		if err != nil {
			return nil, fmt.Errorf("value is out of range, must be positive")
		}
		count = &parsed
	}

	value, err := ctx.Store.RPop(ctx.DBIndex, args[0], count)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return ctx.Store.Protocol.EncodeNil(), nil
	}

	return valueToRESP(value), nil
}
