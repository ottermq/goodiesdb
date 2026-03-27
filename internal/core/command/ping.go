package command

import "github.com/andrelcunha/goodiesdb/internal/protocol"

type PingCommand struct{}

func NewPingCommand() *PingCommand {
	return &PingCommand{}
}

func (c *PingCommand) Name() string {
	return "PING"
}

func (c *PingCommand) MinArgs() int {
	return 0
}

func (c *PingCommand) MaxArgs() int {
	return -1
}

func (c *PingCommand) RequiresAuth() bool {
	return false
}

func (c *PingCommand) Validate(args []string) error {
	return nil
}

func (c *PingCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	if len(args) == 0 {
		return protocol.SimpleString("PONG"), nil
	}
	return protocol.BulkString([]byte(args[0])), nil
}
