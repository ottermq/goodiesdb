package command

import "github.com/ottermq/goodiesdb/internal/protocol"

type DecrCommand struct{}

func NewDecrCommand() *DecrCommand {
	return &DecrCommand{}
}

func (c *DecrCommand) Name() string {
	return "DECR"
}

func (c *DecrCommand) RequiresAuth() bool {
	return false // TODO: refactor authentication
}

func (c *DecrCommand) Validate(args []string) error {
	if len(args) != 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (c *DecrCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	value, err := ctx.Store.Decr(ctx.DBIndex, args[0])
	if err != nil {
		return nil, err
	}
	return protocol.Integer(int64(value)), nil
}
