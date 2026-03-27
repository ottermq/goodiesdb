package command

import "github.com/andrelcunha/goodiesdb/internal/protocol"

type TTLCommand struct{}

func NewTTLCommand() *TTLCommand {
	return &TTLCommand{}
}

func (c *TTLCommand) Name() string {
	return "TTL"
}

func (c *TTLCommand) MinArgs() int {
	return 1
}

func (c *TTLCommand) MaxArgs() int {
	return 1
}

func (c *TTLCommand) RequiresAuth() bool {
	return false // TODO: refactor authentication
}

func (c *TTLCommand) Validate(args []string) error {
	if len(args) != 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (c *TTLCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	ttl, err := ctx.Store.TTL(ctx.DBIndex, args[0])
	if err != nil {
		return nil, err
	}
	return protocol.Integer(int64(ttl)), nil
}
