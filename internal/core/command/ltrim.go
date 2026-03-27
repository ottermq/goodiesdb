package command

import (
	"fmt"
	"strconv"

	"github.com/andrelcunha/goodiesdb/internal/protocol"
)

type LTrimCommand struct{}

func NewLTrimCommand() *LTrimCommand {
	return &LTrimCommand{}
}

func (c *LTrimCommand) Name() string {
	return "LTRIM"
}

func (c *LTrimCommand) RequiresAuth() bool {
	return false // TODO: refactor authentication
}

func (c *LTrimCommand) Validate(args []string) error {
	if len(args) != 3 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (c *LTrimCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	start, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, fmt.Errorf("value is not an integer or out of range")
	}
	stop, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, fmt.Errorf("value is not an integer or out of range")
	}

	if err := ctx.Store.LTrim(ctx.DBIndex, args[0], start, stop); err != nil {
		return nil, err
	}

	return protocol.SimpleString("OK"), nil
}
