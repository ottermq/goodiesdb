package command

import "github.com/andrelcunha/goodiesdb/internal/protocol"

type SetNXCommand struct{}

func NewSetNXCommand() *SetNXCommand {
	return &SetNXCommand{}
}

func (c *SetNXCommand) Name() string {
	return "SETNX"
}

func (c *SetNXCommand) MinArgs() int {
	return 2
}

func (c *SetNXCommand) MaxArgs() int {
	return 2
}

func (c *SetNXCommand) RequiresAuth() bool {
	return false // TODO: refactor authentication
}

func (c *SetNXCommand) Validate(args []string) error {
	if len(args) != 2 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (c *SetNXCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	result := ctx.Store.SetNX(ctx.DBIndex, args[0], args[1])
	return protocol.Integer(result), nil
}
