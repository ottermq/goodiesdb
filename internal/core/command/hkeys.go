package command

import "github.com/ottermq/goodiesdb/internal/protocol"

type HKeysCommand struct{}

func NewHKeysCommand() *HKeysCommand {
	return &HKeysCommand{}
}

func (c *HKeysCommand) Name() string {
	return "HKEYS"
}

func (c *HKeysCommand) RequiresAuth() bool {
	return false
}

func (c *HKeysCommand) Validate(args []string) error {
	return requireExactArgs(args, 1)
}

func (c *HKeysCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	keys, err := ctx.Store.HKeys(ctx.DBIndex, args[0])
	if err != nil {
		return nil, err
	}
	return stringSliceToRESPArray(keys), nil
}
