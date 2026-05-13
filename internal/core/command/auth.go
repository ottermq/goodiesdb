package command

import "github.com/ottermq/goodiesdb/internal/protocol"

type AuthCommand struct{}

func NewAuthCommand() *AuthCommand {
	return &AuthCommand{}
}

func (c *AuthCommand) Name() string {
	return "AUTH"
}

func (c *AuthCommand) RequiresAuth() bool {
	return false
}

func (c *AuthCommand) Validate(args []string) error {
	if len(args) != 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (c *AuthCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	if ctx.Auth != nil && ctx.Auth(args[0]) {
		return protocol.SimpleString("OK"), nil
	}
	return protocol.ErrorString("ERR invalid password"), nil
}
