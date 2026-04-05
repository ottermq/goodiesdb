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

func requireFieldValuePairs(args []string) error {
	if len(args) < 3 || len(args)%2 == 0 {
		return ErrWrongNumberOfArguments
	}
	return nil
}

func requireKeyWithFields(args []string) error {
	return requireMinArgs(args, 2)
}

func hashFieldValueArgs(args []string) map[string]any {
	fields := make(map[string]any, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		fields[args[i]] = args[i+1]
	}
	return fields
}
