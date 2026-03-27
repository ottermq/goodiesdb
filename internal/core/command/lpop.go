package command

import (
	"fmt"
	"strconv"

	"github.com/andrelcunha/goodiesdb/internal/protocol"
)

type LPopCommand struct{}

func NewLPopCommand() *LPopCommand {
	return &LPopCommand{}
}

func (c *LPopCommand) Name() string {
	return "LPOP"
}

func (c *LPopCommand) RequiresAuth() bool {
	return false // TODO: refactor authentication
}

func (c *LPopCommand) Validate(args []string) error {
	if len(args) != 1 && len(args) != 2 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (c *LPopCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	var count *int
	if len(args) == 2 {
		parsed, err := strconv.Atoi(args[1])
		if err != nil {
			return nil, fmt.Errorf("value is out of range, must be positive")
		}
		count = &parsed
	}

	value, err := ctx.Store.LPop(ctx.DBIndex, args[0], count)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return ctx.Store.Protocol.EncodeNil(), nil
	}

	return valueToRESP(value), nil
}
