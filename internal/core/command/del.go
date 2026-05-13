package command

import "github.com/ottermq/goodiesdb/internal/protocol"

type DelCommand struct{}

func NewDelCommand() *DelCommand {
	return &DelCommand{}
}

func (c *DelCommand) Name() string {
	return "DEL"
}

func (c *DelCommand) RequiresAuth() bool {
	return false // TODO: refactor authentication
}

func (c *DelCommand) Validate(args []string) error {
	if len(args) != 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (c *DelCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	ctx.Store.Del(ctx.DBIndex, args[0])
	return protocol.Integer(1), nil
}
