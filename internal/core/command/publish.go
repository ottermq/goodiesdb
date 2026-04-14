package command

import (
	"fmt"

	"github.com/andrelcunha/goodiesdb/internal/protocol"
)

type PublishCommand struct{}

func NewPublishCommand() *PublishCommand {
	return &PublishCommand{}
}

func (c *PublishCommand) Name() string { return "PUBLISH" }

func (c *PublishCommand) RequiresAuth() bool { return false }

func (c *PublishCommand) Validate(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("wrong number of arguments for 'publish' command")
	}
	return nil
}

func (c *PublishCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	channel, message := args[0], args[1]
	n := ctx.PubSub.Publish(channel, message)
	return protocol.Integer(n), nil
}
