package command

import "github.com/ottermq/goodiesdb/internal/protocol"

type ExistsCommand struct{}

func NewExistsCommand() *ExistsCommand {
	return &ExistsCommand{}
}

func (c *ExistsCommand) Name() string {
	return "EXISTS"
}

func (c *ExistsCommand) RequiresAuth() bool {
	return false // TODO: refactor authentication
}

func (c *ExistsCommand) Validate(args []string) error {
	if len(args) < 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (c *ExistsCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	count := ctx.Store.Exists(ctx.DBIndex, args...)
	return protocol.Integer(count), nil
}
