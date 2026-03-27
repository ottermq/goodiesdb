package command

import (
	"fmt"
	"strings"

	"github.com/andrelcunha/goodiesdb/internal/protocol"
)

type ScanCommand struct{}

func NewScanCommand() *ScanCommand {
	return &ScanCommand{}
}

func (c *ScanCommand) Name() string {
	return "SCAN"
}

func (c *ScanCommand) RequiresAuth() bool {
	return false
}

func (c *ScanCommand) Validate(args []string) error {
	return requireMinArgs(args, 1)
}

func (c *ScanCommand) Execute(ctx *Context, args []string) (protocol.RESPValue, error) {
	cursor, err := parseIntArg(args[0], "invalid cursor")
	if err != nil {
		return nil, err
	}

	pattern := "*"
	count := 10

	for i := 1; i < len(args); i++ {
		switch strings.ToUpper(args[i]) {
		case "MATCH":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("syntax error")
			}
			pattern = args[i+1]
			i++
		case "COUNT":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("syntax error")
			}
			c, err := parseIntArg(args[i+1], "value is not an integer or out of range")
			if err != nil || c <= 0 {
				if err != nil {
					return nil, err
				}
				return nil, fmt.Errorf("value is not an integer or out of range")
			}
			count = c
			i++
		default:
			return nil, fmt.Errorf("syntax error")
		}
	}

	newCursor, keys, err := ctx.Store.Scan(ctx.DBIndex, cursor, pattern, count)
	if err != nil {
		return nil, err
	}

	return protocol.Array{
		protocol.BulkString([]byte(fmt.Sprintf("%d", newCursor))),
		stringSliceToRESPArray(keys),
	}, nil
}
