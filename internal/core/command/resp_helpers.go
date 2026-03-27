package command

import (
	"fmt"

	"github.com/andrelcunha/goodiesdb/internal/protocol"
)

func valueToRESP(value any) protocol.RESPValue {
	switch v := value.(type) {
	case string:
		return protocol.BulkString([]byte(v))
	case []any:
		return anySliceToRESPArray(v)
	default:
		return protocol.BulkString([]byte(fmt.Sprintf("%v", v)))
	}
}

func anySliceToRESPArray(items []any) protocol.Array {
	arr := make(protocol.Array, len(items))
	for i, item := range items {
		arr[i] = valueToRESP(item)
	}
	return arr
}
