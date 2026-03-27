package command

import "github.com/andrelcunha/goodiesdb/internal/protocol"

type TypeCommand struct{}

func NewTypeCommand() *TypeCommand {
	return &TypeCommand{}
}

func (c *TypeCommand) Name() string {
	return "TYPE"
}

func (c *TypeCommand) RequiresAuth() bool {
	return false // TODO: refactor authentication
}

func (c *TypeCommand) Validate(args []string) error {
	if len(args) != 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (c *TypeCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	return protocol.SimpleString(ctx.Store.Type(ctx.DBIndex, args[0])), nil
}
