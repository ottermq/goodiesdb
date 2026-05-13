package command

import "github.com/ottermq/goodiesdb/internal/protocol"

type HLenCommand struct{}

func NewHLenCommand() *HLenCommand {
	return &HLenCommand{}
}

func (c *HLenCommand) Name() string {
	return "HLEN"
}

func (c *HLenCommand) RequiresAuth() bool {
	return false
}

func (c *HLenCommand) Validate(args []string) error {
	return requireExactArgs(args, 1)
}

func (c *HLenCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	length, err := ctx.Store.HLen(ctx.DBIndex, args[0])
	if err != nil {
		return nil, err
	}
	return protocol.Integer(int64(length)), nil
}
