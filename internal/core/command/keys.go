package command

import "github.com/andrelcunha/goodiesdb/internal/protocol"

type KeysCommand struct{}

func NewKeysCommand() *KeysCommand {
	return &KeysCommand{}
}

func (c *KeysCommand) Name() string {
	return "KEYS"
}

func (c *KeysCommand) MinArgs() int {
	return 1
}

func (c *KeysCommand) MaxArgs() int {
	return 1
}

func (c *KeysCommand) RequiresAuth() bool {
	return false
}

func (c *KeysCommand) Validate(args []string) error {
	if len(args) != 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (c *KeysCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	keys, err := ctx.Store.Keys(ctx.DBIndex, args[0])
	if err != nil {
		return nil, err
	}
	return stringSliceToRESPArray(keys), nil
}
