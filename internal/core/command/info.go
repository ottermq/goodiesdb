package command

import "github.com/andrelcunha/goodiesdb/internal/protocol"

type InfoCommand struct{}

func NewInfoCommand() *InfoCommand {
	return &InfoCommand{}
}

func (c *InfoCommand) Name() string {
	return "INFO"
}

func (c *InfoCommand) MinArgs() int {
	return 0
}

func (c *InfoCommand) MaxArgs() int {
	return -1
}

func (c *InfoCommand) RequiresAuth() bool {
	return false
}

func (c *InfoCommand) Validate(args []string) error {
	return nil
}

func (c *InfoCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	if ctx.Info == nil {
		return protocol.BulkString([]byte("")), nil
	}
	return protocol.BulkString([]byte(ctx.Info())), nil
}
