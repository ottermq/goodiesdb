package command

import "github.com/andrelcunha/goodiesdb/internal/protocol"

type QuitCommand struct{}

func NewQuitCommand() *QuitCommand {
	return &QuitCommand{}
}

func (c *QuitCommand) Name() string {
	return "QUIT"
}

func (c *QuitCommand) MinArgs() int {
	return 0
}

func (c *QuitCommand) MaxArgs() int {
	return -1
}

func (c *QuitCommand) RequiresAuth() bool {
	return false
}

func (c *QuitCommand) Validate(args []string) error {
	return nil
}

func (c *QuitCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	return protocol.SimpleString("OK"), nil
}
