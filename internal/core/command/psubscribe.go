package command

import (
	"fmt"

	"github.com/andrelcunha/goodiesdb/internal/protocol"
)

type PSubscribeCommand struct{}

func NewPSubscribeCommand() *PSubscribeCommand {
	return &PSubscribeCommand{}
}

func (c *PSubscribeCommand) Name() string { return "PSUBSCRIBE" }

func (c *PSubscribeCommand) RequiresAuth() bool { return false }

func (c *PSubscribeCommand) Validate(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("wrong number of arguments for 'psubscribe' command")
	}
	return nil
}

func (c *PSubscribeCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	for _, pattern := range args {
		count := ctx.PubSub.PSubscribe(ctx.Conn, pattern)
		ctx.Write(protocol.Array{
			protocol.BulkString("psubscribe"),
			protocol.BulkString(pattern),
			protocol.Integer(count),
		})
	}
	ctx.SetMode(int(modeSubscriber))
	return nil, nil
}
