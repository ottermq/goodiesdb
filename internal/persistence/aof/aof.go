package aof

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strconv"

	"github.com/andrelcunha/goodiesdb/internal/core/store"
	"github.com/andrelcunha/goodiesdb/internal/logging"
	"github.com/andrelcunha/goodiesdb/internal/protocol"
	"github.com/andrelcunha/goodiesdb/internal/protocol/resp2"
)

// AOFWriter writes commands to a file
func AOFWriter(aofChan chan store.AOFCommand, filename string) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logging.Errorf("Failed to open AOF file: %v", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	proto := &resp2.RESP2Protocol{}

	for cmd := range aofChan {
		if err := proto.Encode(writer, toRESPArray(cmd)); err != nil {
			logging.Errorf("Failed to write to AOF file: %v", err)
			return
		}
		if err := writer.Flush(); err != nil {
			logging.Errorf("Failed to flush AOF file: %v", err)
			return
		}
	}
}

// RebuildStoreFromAOF rebuilds the store from the AOF file
func RebuildStoreFromAOF(s *store.Store, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	peek, err := reader.Peek(1)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	if peek[0] != '*' {
		logging.Infof("Legacy line-based AOF detected in %s; ignoring file and starting with an empty store", filename)
		return nil
	}

	proto := &resp2.RESP2Protocol{}
	for {
		value, err := proto.Parse(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		parts, ok := respArrayToCommand(value)
		if !ok || len(parts) == 0 {
			continue
		}

		dispatchAOF(parts, s)
	}
}

func toRESPArray(cmd store.AOFCommand) protocol.Array {
	arr := make(protocol.Array, len(cmd))
	for i, part := range cmd {
		arr[i] = protocol.BulkString([]byte(part))
	}
	return arr
}

func respArrayToCommand(value protocol.RESPValue) ([]string, bool) {
	arr, ok := value.(protocol.Array)
	if !ok {
		return nil, false
	}

	parts := make([]string, len(arr))
	for i, part := range arr {
		switch v := part.(type) {
		case protocol.BulkString:
			parts[i] = string(v)
		case protocol.SimpleString:
			parts[i] = string(v)
		default:
			return nil, false
		}
	}

	return parts, true
}

func dispatchAOF(parts []string, s *store.Store) {
	cmdName := parts[0]
	dbIndex, ok := parseDBIndex(parts)
	if !ok && cmdName != "FLUSHALL" {
		logging.Errorf("Invalid AOF command database index for %q", cmdName)
		return
	}

	switch cmdName {
	case "SET":
		aofSet(parts, s, dbIndex)
	case "DEL":
		aofDel(parts, s, dbIndex)
	case "SETNX":
		aofSetNX(parts, s, dbIndex)
	case "EXPIRE":
		aofExpire(parts, s, dbIndex)
	case "INCR":
		aofIncr(parts, s, dbIndex)
	case "DECR":
		aofDecr(parts, s, dbIndex)
	case "LPUSH":
		aofLPush(parts, s, dbIndex)
	case "RPUSH":
		aofRPush(parts, s, dbIndex)
	case "LPOP":
		aofLPop(parts, s, dbIndex)
	case "RPOP":
		aofRpop(parts, s, dbIndex)
	case "LTRIM":
		aofLTrim(parts, s, dbIndex)
	case "RENAME":
		aofRename(parts, s, dbIndex)
	case "FLUSHDB":
		aofFlushDB(parts, s, dbIndex)
	case "FLUSHALL":
		aofFlushAll(parts, s)
	case "HSET":
		aofHSet(parts, s, dbIndex)
	case "HDEL":
		aofHDel(parts, s, dbIndex)
	default:
		logging.Errorf("Unknown AOF command: %v", parts)
	}
}

func parseDBIndex(parts []string) (int, bool) {
	if len(parts) < 2 {
		return 0, false
	}
	dbIndex, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, false
	}
	return dbIndex, true
}
