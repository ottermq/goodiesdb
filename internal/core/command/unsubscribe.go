package command

import "github.com/andrelcunha/goodiesdb/internal/protocol"

type UnsubscribeCommand struct{}

func NewUnsubscribeCommand() *UnsubscribeCommand {
	return &UnsubscribeCommand{}
}

func (c *UnsubscribeCommand) Name() string { return "UNSUBSCRIBE" }

func (c *UnsubscribeCommand) RequiresAuth() bool { return false }

func (c *UnsubscribeCommand) Validate(args []string) error { return nil }

func (c *UnsubscribeCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	// Zero args means unsubscribe from all — but we don't have a list of current
	// subscriptions in the command layer. Send a single confirmation with nil channel.
	if len(args) == 0 {
		ctx.PubSub.UnsubscribeAll(ctx.Conn)
		ctx.Write(protocol.Array{
			protocol.BulkString("unsubscribe"),
			protocol.BulkString(nil),
			protocol.Integer(0),
		})
		ctx.SetMode(modeNormal)
		return nil, nil
	}

	for _, channel := range args {
		remaining := ctx.PubSub.Unsubscribe(ctx.Conn, channel)
		ctx.Write(protocol.Array{
			protocol.BulkString("unsubscribe"),
			protocol.BulkString(channel),
			protocol.Integer(remaining),
		})
		if remaining == 0 {
			ctx.SetMode(modeNormal)
		}
	}
	return nil, nil
}

const modeNormal = 0
