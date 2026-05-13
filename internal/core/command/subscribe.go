package command

import (
	"fmt"

	"github.com/ottermq/goodiesdb/internal/protocol"
)

type SubscribeCommand struct{}

func NewSubscribeCommand() *SubscribeCommand {
	return &SubscribeCommand{}
}

func (c *SubscribeCommand) Name() string { return "SUBSCRIBE" }

func (c *SubscribeCommand) RequiresAuth() bool { return false }

func (c *SubscribeCommand) Validate(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("wrong number of arguments for 'subscribe' command")
	}
	return nil
}

func (c *SubscribeCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	for _, channel := range args {
		count := ctx.PubSub.Subscribe(ctx.Conn, channel)
		ctx.Write(protocol.Array{
			protocol.BulkString("subscribe"),
			protocol.BulkString(channel),
			protocol.Integer(count),
		})
	}
	// Flip the connection to subscriber mode. The server will start the write
	// goroutine that drains the broker's delivery channel for this connection.
	ctx.SetMode(int(modeSubscriber))
	return nil, nil
}

// modeSubscriber mirrors server.ModeSubscriber without creating an import cycle.
const modeSubscriber = 1
