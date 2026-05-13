package command

import "github.com/ottermq/goodiesdb/internal/protocol"

type IncrCommand struct{}

func NewIncrCommand() *IncrCommand {
	return &IncrCommand{}
}

func (c *IncrCommand) Name() string {
	return "INCR"
}

func (c *IncrCommand) RequiresAuth() bool {
	return false // TODO: refactor authentication
}

func (c *IncrCommand) Validate(args []string) error {
	if len(args) != 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (c *IncrCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	value, err := ctx.Store.Incr(ctx.DBIndex, args[0])
	if err != nil {
		return nil, err
	}
	return protocol.Integer(int64(value)), nil
}
