package command

import "github.com/ottermq/goodiesdb/internal/protocol"

type HGetCommand struct{}

func NewHGetCommand() *HGetCommand {
	return &HGetCommand{}
}

func (c *HGetCommand) Name() string {
	return "HGET"
}

func (c *HGetCommand) RequiresAuth() bool {
	return false
}

func (c *HGetCommand) Validate(args []string) error {
	return requireExactArgs(args, 2)
}

func (c *HGetCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	value, ok, err := ctx.Store.HGet(ctx.DBIndex, args[0], args[1])
	if err != nil {
		return nil, err
	}
	if !ok {
		return ctx.Nil(), nil
	}
	return protocol.BulkString([]byte(value)), nil
}
