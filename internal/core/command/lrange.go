package command

import (
	"fmt"

	"github.com/ottermq/goodiesdb/internal/protocol"
)

type LRangeCommand struct{}

func NewLRangeCommand() *LRangeCommand {
	return &LRangeCommand{}
}

func (c *LRangeCommand) Name() string {
	return "LRANGE"
}

func (c *LRangeCommand) RequiresAuth() bool {
	return false // TODO: refactor authentication
}

func (c *LRangeCommand) Validate(args []string) error {
	return requireExactArgs(args, 3)
}

func (c *LRangeCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	start, err := parseIntArg(args[1], "value is not an integer or out of range")
	if err != nil {
		return nil, err
	}
	stop, err := parseIntArg(args[2], "value is not an integer or out of range")
	if err != nil {
		return nil, err
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
