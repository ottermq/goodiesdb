package command

import (
	"fmt"
	"sort"

	"github.com/ottermq/goodiesdb/internal/protocol"
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

func stringSliceToRESPArray(items []string) protocol.Array {
	arr := make(protocol.Array, len(items))
	for i, item := range items {
		arr[i] = protocol.BulkString([]byte(item))
	}
	return arr
}

func hashToRESPArray(items map[string]string) protocol.Array {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	arr := make(protocol.Array, 0, len(items)*2)
	for _, key := range keys {
		arr = append(arr, protocol.BulkString([]byte(key)))
		arr = append(arr, protocol.BulkString([]byte(items[key])))
	}
	return arr
}
