package command

import (
	"strings"

	"github.com/andrelcunha/goodiesdb/internal/protocol"
)

type HelloCommand struct{}

func NewHelloCommand() *HelloCommand {
	return &HelloCommand{}
}

func (c *HelloCommand) Name() string {
	return "HELLO"
}

func (c *HelloCommand) RequiresAuth() bool {
	return false
}

func (c *HelloCommand) Validate(args []string) error {
	if len(args) > 1 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (c *HelloCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	if len(args) == 0 {
		return c.ReturnHello2(ctx), nil

	}
	switch strings.ToUpper(args[0]) {
	case "2":
		return c.ReturnHello2(ctx), nil
	case "3":
		return protocol.ErrorString("NOPROTO this server does not support RESP3"), nil
	default:
		return protocol.ErrorString("ERR Protocol version is not supported or command arguments are not correct."), nil
	}
}

func (c *HelloCommand) ReturnHello2(ctx *Context) protocol.RESPValue {
	return protocol.Array{
		protocol.BulkString("server"), protocol.BulkString("goodiesdb"),
		protocol.BulkString("version"), protocol.BulkString(ctx.GetVersion()),
		protocol.BulkString("proto"), protocol.Integer(ctx.Protocol.ProtocolVersion()),
		protocol.BulkString("id"), protocol.Integer(ctx.GetConnID()),
		protocol.BulkString("mode"), protocol.BulkString("standalone"),
		protocol.BulkString("role"), protocol.BulkString("master"),
		protocol.BulkString("modules"), protocol.Array{},
	}
}
