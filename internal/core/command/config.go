package command

import (
	"strings"

	"github.com/ottermq/goodiesdb/internal/protocol"
)

type ConfigCommand struct{}

func NewConfigCommand() *ConfigCommand {
	return &ConfigCommand{}
}

func (c *ConfigCommand) Name() string {
	return "CONFIG"
}

func (c *ConfigCommand) RequiresAuth() bool {
	return false
}

func (c *ConfigCommand) Validate(args []string) error {
	if len(args) < 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (c *ConfigCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	switch strings.ToUpper(args[0]) {
	case "GET":
		return protocol.Array{}, nil
	case "SET":
		return protocol.SimpleString("OK"), nil
	case "RESETSTAT":
		return protocol.SimpleString("OK"), nil
	default:
		return protocol.ErrorString("ERR unknown subcommand '" + args[0] + ".'"), nil
	}
}
