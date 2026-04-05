package command

import "github.com/andrelcunha/goodiesdb/internal/protocol"

type HDelCommand struct{}

func NewHDelCommand() *HDelCommand {
	return &HDelCommand{}
}

func (c *HDelCommand) Name() string {
	return "HDEL"
}

func (c *HDelCommand) RequiresAuth() bool {
	return false
}

func (c *HDelCommand) Validate(args []string) error {
	return requireKeyWithFields(args)
}

func (c *HDelCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	deleted, err := ctx.Store.HDel(ctx.DBIndex, args[0], args[1:]...)
	if err != nil {
		return nil, err
	}
	return protocol.Integer(int64(deleted)), nil
}
