package command

import "github.com/andrelcunha/goodiesdb/internal/protocol"

type HGetAllCommand struct{}

func NewHGetAllCommand() *HGetAllCommand {
	return &HGetAllCommand{}
}

func (c *HGetAllCommand) Name() string {
	return "HGETALL"
}

func (c *HGetAllCommand) RequiresAuth() bool {
	return false
}

func (c *HGetAllCommand) Validate(args []string) error {
	return requireExactArgs(args, 1)
}

func (c *HGetAllCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	values, err := ctx.Store.HGetAll(ctx.DBIndex, args[0])
	if err != nil {
		return nil, err
	}
	return hashToRESPArray(values), nil
}
