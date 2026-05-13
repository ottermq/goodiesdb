package command

import "github.com/ottermq/goodiesdb/internal/protocol"

type RenameCommand struct{}

func NewRenameCommand() *RenameCommand {
	return &RenameCommand{}
}

func (c *RenameCommand) Name() string {
	return "RENAME"
}

func (c *RenameCommand) RequiresAuth() bool {
	return false
}

func (c *RenameCommand) Validate(args []string) error {
	if len(args) != 2 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (c *RenameCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	if err := ctx.Store.Rename(ctx.DBIndex, args[0], args[1]); err != nil {
		return nil, err
	}
	return protocol.SimpleString("OK"), nil
}
