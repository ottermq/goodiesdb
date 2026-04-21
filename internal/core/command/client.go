package command

import (
	"strings"

	"github.com/andrelcunha/goodiesdb/internal/protocol"
)

type ClientCommand struct{}

func NewClientCommand() *ClientCommand {
	return &ClientCommand{}
}

func (c *ClientCommand) Name() string {
	return "CLIENT"
}

func (c *ClientCommand) RequiresAuth() bool {
	return false
}

func (c *ClientCommand) Validate(args []string) error {
	if len(args) < 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (c *ClientCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	switch strings.ToUpper(args[0]) {
	case "SETNAME":
		if len(args) != 2 {
			return protocol.ErrorString("ERR syntax error."), nil
		}
		if strings.ContainsAny(args[1], " \n\r") {
			return protocol.ErrorString("ERR Client names cannot contain spaces, newlines or special characters."), nil
		}
		ctx.SetConnName(args[1])
		return protocol.SimpleString("OK"), nil

	case "GETNAME":
		name := ctx.GetConnName()
		if name == "" {
			return ctx.Nil(), nil
		}
		return protocol.BulkString([]byte(name)), nil

	case "ID":
		return protocol.Integer(ctx.GetConnID()), nil

	case "INFO":
		if ctx.GetConnInfo == nil {
			return protocol.BulkString([]byte("")), nil
		}
		return protocol.BulkString([]byte(ctx.GetConnInfo())), nil

	case "LIST":
		// same as INFO for now
		if ctx.GetConnInfo == nil {
			return protocol.BulkString([]byte("")), nil
		}
		return protocol.BulkString([]byte(ctx.GetConnInfo())), nil

	case "NO-EVICT", "NO-TOUCH":
		// no-ops for now
		if len(args) != 2 {
			return protocol.ErrorString("ERR syntax error"), nil
		}
		switch strings.ToUpper(args[1]) {
		case "ON", "OFF":
			return protocol.SimpleString("OK"), nil
		default:
			return protocol.ErrorString("ERR syntax error"), nil
		}

	default:
		return protocol.ErrorString("ERR unknown subcommand '" + args[0] + ".'"), nil
	}
}
