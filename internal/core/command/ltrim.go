package command

import "github.com/andrelcunha/goodiesdb/internal/protocol"

type LTrimCommand struct{}

func NewLTrimCommand() *LTrimCommand {
	return &LTrimCommand{}
}

func (c *LTrimCommand) Name() string {
	return "LTRIM"
}

func (c *LTrimCommand) RequiresAuth() bool {
	return false // TODO: refactor authentication
}

func (c *LTrimCommand) Validate(args []string) error {
	return requireExactArgs(args, 3)
}

func (c *LTrimCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	start, err := parseIntArg(args[1], "value is not an integer or out of range")
	if err != nil {
		return nil, err
	}
	stop, err := parseIntArg(args[2], "value is not an integer or out of range")
	if err != nil {
		return nil, err
	}

	if err := ctx.Store.LTrim(ctx.DBIndex, args[0], start, stop); err != nil {
		return nil, err
	}

	return protocol.SimpleString("OK"), nil
}
