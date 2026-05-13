package command

import "github.com/ottermq/goodiesdb/internal/protocol"

type PUnsubscribeCommand struct{}

func NewPUnsubscribeCommand() *PUnsubscribeCommand {
	return &PUnsubscribeCommand{}
}

func (c *PUnsubscribeCommand) Name() string { return "PUNSUBSCRIBE" }

func (c *PUnsubscribeCommand) RequiresAuth() bool { return false }

func (c *PUnsubscribeCommand) Validate(args []string) error { return nil }

func (c *PUnsubscribeCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	if len(args) == 0 {
		ctx.PubSub.UnsubscribeAll(ctx.Conn)
		ctx.Write(protocol.Array{
			protocol.BulkString("punsubscribe"),
			protocol.BulkString(nil),
			protocol.Integer(0),
		})
		ctx.SetMode(modeNormal)
		return nil, nil
	}

	for _, pattern := range args {
		remaining := ctx.PubSub.PUnsubscribe(ctx.Conn, pattern)
		ctx.Write(protocol.Array{
			protocol.BulkString("punsubscribe"),
			protocol.BulkString(pattern),
			protocol.Integer(remaining),
		})
		if remaining == 0 {
			ctx.SetMode(modeNormal)
		}
	}
	return nil, nil
}
