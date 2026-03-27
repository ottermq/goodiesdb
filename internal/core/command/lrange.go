package command

import (
	"fmt"
	"strconv"

	"github.com/andrelcunha/goodiesdb/internal/protocol"
)

type LRangeCommand struct{}

func NewLRangeCommand() *LRangeCommand {
	return &LRangeCommand{}
}

func (c *LRangeCommand) Name() string {
	return "LRANGE"
}

func (c *LRangeCommand) MinArgs() int {
	return 3
}

func (c *LRangeCommand) MaxArgs() int {
	return 3
}

func (c *LRangeCommand) RequiresAuth() bool {
	return false // TODO: refactor authentication
}

func (c *LRangeCommand) Validate(args []string) error {
	if len(args) != 3 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (c *LRangeCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	start, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, fmt.Errorf("value is not an integer or out of range")
	}
	stop, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, fmt.Errorf("value is not an integer or out of range")
	}

	values, err := ctx.Store.LRange(ctx.DBIndex, args[0], start, stop)
	if err != nil {
		return nil, err
	}

	result := make(protocol.Array, len(values))
	for i, value := range values {
		result[i] = protocol.BulkString([]byte(fmt.Sprintf("%v", value)))
	}

	return result, nil
}
