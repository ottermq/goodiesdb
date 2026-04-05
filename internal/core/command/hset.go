package command

import "github.com/andrelcunha/goodiesdb/internal/protocol"

type HSetCommand struct{}

func NewHSetCommand() *HSetCommand {
	return &HSetCommand{}
}

func (c *HSetCommand) Name() string {
	return "HSET"
}

func (c *HSetCommand) RequiresAuth() bool {
	return false
}

func (c *HSetCommand) Validate(args []string) error {
	return requireFieldValuePairs(args)
}

func (c *HSetCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	added, err := ctx.Store.HSet(ctx.DBIndex, args[0], hashFieldValueArgs(args[1:]))
	if err != nil {
		return nil, err
	}
	return protocol.Integer(int64(added)), nil
}
