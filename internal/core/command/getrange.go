package command

import "github.com/andrelcunha/goodiesdb/internal/protocol"

type GetRangeCommand struct{}

func NewGetRangeCommand() *GetRangeCommand {
	return &GetRangeCommand{}
}

func (c *GetRangeCommand) Name() string {
	return "GETRANGE"
}

func (c *GetRangeCommand) RequiresAuth() bool {
	return false // TODO: refactor authentication
}

func (c *GetRangeCommand) Validate(args []string) error {
	return requireExactArgs(args, 3)
}

func (c *GetRangeCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	start, err := parseIntArg(args[1], "value is not an integer or out of range")
	if err != nil {
		return nil, err
	}
	end, err := parseIntArg(args[2], "value is not an integer or out of range")
	if err != nil {
		return nil, err
	}

	value, err := ctx.Store.GetRange(ctx.DBIndex, args[0], start, end)
	if err != nil {
		return nil, err
	}

	return protocol.BulkString([]byte(value)), nil
}
