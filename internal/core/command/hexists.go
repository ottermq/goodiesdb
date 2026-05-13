package command

import "github.com/ottermq/goodiesdb/internal/protocol"

type HExistsCommand struct{}

func NewHExistsCommand() *HExistsCommand {
	return &HExistsCommand{}
}

func (c *HExistsCommand) Name() string {
	return "HEXISTS"
}

func (c *HExistsCommand) RequiresAuth() bool {
	return false
}

func (c *HExistsCommand) Validate(args []string) error {
	return requireExactArgs(args, 2)
}

func (c *HExistsCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	exists, err := ctx.Store.HExists(ctx.DBIndex, args[0], args[1])
	if err != nil {
		return nil, err
	}
	if exists {
		return protocol.Integer(1), nil
	}
	return protocol.Integer(0), nil
}
