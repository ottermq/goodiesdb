package command

import "github.com/andrelcunha/goodiesdb/internal/protocol"

type StrLenCommand struct{}

func NewStrLenCommand() *StrLenCommand {
	return &StrLenCommand{}
}

func (c *StrLenCommand) Name() string {
	return "STRLEN"
}

func (c *StrLenCommand) MinArgs() int {
	return 1
}

func (c *StrLenCommand) MaxArgs() int {
	return 1
}

func (c *StrLenCommand) RequiresAuth() bool {
	return false // TODO: refactor authentication
}

func (c *StrLenCommand) Validate(args []string) error {
	if len(args) != 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (c *StrLenCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	length, err := ctx.Store.StrLen(ctx.DBIndex, args[0])
	if err != nil {
		return nil, err
	}
	return protocol.Integer(int64(length)), nil
}
