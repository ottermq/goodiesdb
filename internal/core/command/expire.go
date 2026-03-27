package command

import (
	"fmt"
	"strconv"
	"time"

	"github.com/andrelcunha/goodiesdb/internal/protocol"
)

type ExpireCommand struct{}

func NewExpireCommand() *ExpireCommand {
	return &ExpireCommand{}
}

func (c *ExpireCommand) Name() string {
	return "EXPIRE"
}

func (c *ExpireCommand) RequiresAuth() bool {
	return false // TODO: refactor authentication
}

func (c *ExpireCommand) Validate(args []string) error {
	if len(args) != 2 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func (c *ExpireCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	ttl, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, fmt.Errorf("invalid TTL")
	}

	if ctx.Store.Expire(ctx.DBIndex, args[0], time.Duration(ttl)*time.Second) {
		return protocol.Integer(1), nil
	}

	return protocol.Integer(0), nil
}
