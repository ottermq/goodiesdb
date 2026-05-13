package command

import "github.com/ottermq/goodiesdb/internal/protocol"

type SetCommand struct{}

func NewSetCommand() *SetCommand {
	return &SetCommand{}
}

func (c *SetCommand) Name() string {
	return "SET"
}

func (c *SetCommand) RequiresAuth() bool {
	return false // TODO: refactor authentication
}

func (c *SetCommand) Validate(args []string) error {
	if len(args) < 2 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (c *SetCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	key := args[0]
	value := args[1]
	options := args[2:]
	ok, err := ctx.Store.Set(ctx.DBIndex, key, value, options...)
	if err != nil {
		return nil, err
	}
	if !ok {
		return ctx.Nil(), nil
	}
	return protocol.SimpleString("OK"), nil
}
