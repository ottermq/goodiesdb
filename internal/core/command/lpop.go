package command

import "github.com/andrelcunha/goodiesdb/internal/protocol"

type LPopCommand struct{}

func NewLPopCommand() *LPopCommand {
	return &LPopCommand{}
}

func (c *LPopCommand) Name() string {
	return "LPOP"
}

func (c *LPopCommand) RequiresAuth() bool {
	return false // TODO: refactor authentication
}

func (c *LPopCommand) Validate(args []string) error {
	return requireOneOfArgCounts(args, 1, 2)
}

func (c *LPopCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	var count *int
	if len(args) == 2 {
		parsed, err := parseIntArg(args[1], "value is out of range, must be positive")
		if err != nil {
			return nil, err
		}
		count = &parsed
	}

	value, err := ctx.Store.LPop(ctx.DBIndex, args[0], count)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return ctx.Nil(), nil
	}

	return valueToRESP(value), nil
}
