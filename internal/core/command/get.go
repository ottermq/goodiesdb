package command

import (
	"github.com/andrelcunha/goodiesdb/internal/core/store"
	"github.com/andrelcunha/goodiesdb/internal/protocol"
)

type GetCommand struct{}

func NewGetCommand() *GetCommand {
	return &GetCommand{}
}

func (c *GetCommand) Name() string {
	return "GET"
}

func (c *GetCommand) RequiresAuth() bool {
	return false // TODO: refactor authentication
}

func (c *GetCommand) Validate(args []string) error {
	if len(args) != 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (c *GetCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	value, ok := ctx.Store.Get(ctx.DBIndex, args[0])
	if !ok {
		return ctx.Nil(), nil
	}

	if value.IsExpired() {
		// lazy delete
		ctx.Store.Delete(ctx.DBIndex, args[0])
		return ctx.Nil(), nil
	}

	str, err := value.ToString()
	if err != nil {
		return nil, store.ErrWrongType
	}

	return protocol.BulkString([]byte(str)), nil
}
