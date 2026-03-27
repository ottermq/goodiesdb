package command

import (
	"fmt"
	"strconv"
)

func requireExactArgs(args []string, expected int) error {
	if len(args) != expected {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func requireMinArgs(args []string, min int) error {
	if len(args) < min {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func requireOneOfArgCounts(args []string, counts ...int) error {
	for _, count := range counts {
		if len(args) == count {
			return nil
		}
	}
	return ErrWrongNumberOfArguments
}

func parseIntArg(value string, errMsg string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s", errMsg)
	}
	return parsed, nil
}
