package command

import "github.com/andrelcunha/goodiesdb/internal/protocol"

type HMGetCommand struct{}

func NewHMGetCommand() *HMGetCommand {
	return &HMGetCommand{}
}

func (c *HMGetCommand) Name() string {
	return "HMGET"
}

func (c *HMGetCommand) RequiresAuth() bool {
	return false
}

func (c *HMGetCommand) Validate(args []string) error {
	return requireKeyWithFields(args)
}

func (c *HMGetCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	values, err := ctx.Store.HMGet(ctx.DBIndex, args[0], args[1:]...)
	if err != nil {
		return nil, err
	}

	arr := make(protocol.Array, len(values))
	for i, value := range values {
		if value == nil {
			arr[i] = ctx.Nil()
			continue
		}
		arr[i] = protocol.BulkString([]byte(value.(string)))
	}

	return arr, nil
}
