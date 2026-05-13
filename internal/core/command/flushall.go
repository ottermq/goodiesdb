package command

import "github.com/ottermq/goodiesdb/internal/protocol"

type FlushAllCommand struct{}

func NewFlushAllCommand() *FlushAllCommand {
	return &FlushAllCommand{}
}

func (c *FlushAllCommand) Name() string {
	return "FLUSHALL"
}

func (c *FlushAllCommand) RequiresAuth() bool {
	return false
}

func (c *FlushAllCommand) Validate(args []string) error {
	return nil
}

func (c *FlushAllCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	ctx.Store.FlushAll()
	return protocol.SimpleString("OK"), nil
}
