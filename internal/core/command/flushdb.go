package command

import "github.com/andrelcunha/goodiesdb/internal/protocol"

type FlushDBCommand struct{}

func NewFlushDBCommand() *FlushDBCommand {
	return &FlushDBCommand{}
}

func (c *FlushDBCommand) Name() string {
	return "FLUSHDB"
}

func (c *FlushDBCommand) MinArgs() int {
	return 0
}

func (c *FlushDBCommand) MaxArgs() int {
	return -1
}

func (c *FlushDBCommand) RequiresAuth() bool {
	return false
}

func (c *FlushDBCommand) Validate(args []string) error {
	return nil
}

func (c *FlushDBCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	ctx.Store.FlushDb(ctx.DBIndex)
	return protocol.SimpleString("OK"), nil
}
