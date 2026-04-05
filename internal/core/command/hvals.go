package command

import "github.com/andrelcunha/goodiesdb/internal/protocol"

type HValsCommand struct{}

func NewHValsCommand() *HValsCommand {
	return &HValsCommand{}
}

func (c *HValsCommand) Name() string {
	return "HVALS"
}

func (c *HValsCommand) RequiresAuth() bool {
	return false
}

func (c *HValsCommand) Validate(args []string) error {
	return requireExactArgs(args, 1)
}

func (c *HValsCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	values, err := ctx.Store.HVals(ctx.DBIndex, args[0])
	if err != nil {
		return nil, err
	}
	return stringSliceToRESPArray(values), nil
}
