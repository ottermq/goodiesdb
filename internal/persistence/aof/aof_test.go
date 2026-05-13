package aof

import (
	"bufio"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ottermq/goodiesdb/internal/core/store"
	"github.com/ottermq/goodiesdb/internal/protocol"
	"github.com/ottermq/goodiesdb/internal/protocol/resp2"
	"github.com/ottermq/goodiesdb/internal/utils/slice"
)

func TestRebuildStoreFromAOF(t *testing.T) {
	aofFilename := "test_appendonly.aof"
	os.Remove(aofFilename)
	aofChan := make(chan store.AOFCommand, 100)

	// Start the AOF writer
	go AOFWriter(aofChan, aofFilename)

	// Initialize the store with AOF logging
	s := store.NewStore(aofChan)

	dbIndex := 0

	// Set and expire commands
	s.Set(dbIndex, "Key1", "Value1")
	s.Set(dbIndex, "Key2", "Value2")
	s.Expire(dbIndex, "Key1", 3*time.Second)

	// SETNX command
	s.SetNX(dbIndex, "Key3", "Value3")   // Should succeed
	s.SetNX(dbIndex, "Key3", "NewValue") // Should fail because Key3 already exists

	// List commands
	s.LPush(dbIndex, "List1", "Value1", "Value2", "Value3")
	s.RPush(dbIndex, "List1", "Value4")
	s.LPop(dbIndex, "List1", nil)
	s.RPop(dbIndex, "List1", nil)

	// List trimming commands
	s.LTrim(dbIndex, "List1", 1, 2)

	// Rename command
	s.Rename(dbIndex, "Key2", "RenamedKey")

	// Give some time for commands to be written to AOF
	time.Sleep(1 * time.Second)

	// Rebuild state from AOF
	newAofFilename := "new_test_appendonly.aof"
	os.Remove(newAofFilename)
	newAofChan := make(chan store.AOFCommand, 100)
	go AOFWriter(newAofChan, newAofFilename)

	newStore := store.NewStore(newAofChan)

	err := RebuildStoreFromAOF(newStore, aofFilename)
	if err != nil {
		t.Fatalf("Failed to rebuild state from AOF: %v", err)
	}
	// set new aofFilename

	// Verify Key2 has been renamed to RenamedKey
	value, ok := newStore.Get(dbIndex, "RenamedKey")
	valStr := value.Data.(string)

	if !ok || valStr != "Value2" {
		t.Errorf("Expected Value2 for RenamedKey, got %s", valStr)
		t.Fail()
	}

	// Verify List1 contents
	list, _ := newStore.LRange(dbIndex, "List1", 0, -1)
	expectedList := []string{"Value1"}
	listStr := make([]string, len(list))
	for i, v := range list {
		listStr[i] = v.(string)
	}
	if !slice.Equal(listStr, expectedList) {
		t.Errorf("Expected %v, got %v", expectedList, list)
		t.Fail()
	}

	// Wait for the key to expire
	time.Sleep(4 * time.Second)

	// Verify Key1 exists after it expires
	if newStore.Exists(dbIndex, "Key1") > 0 {
		t.Errorf("Expected Key1 to be expired after waiting more than 3 seconds")
		t.Fail()
	}

	// Clean up the AOF file
	os.Remove(aofFilename)
	os.Remove(newAofFilename)
}

// Test aofRename
func TestAofRename(t *testing.T) {
	cmd := "RENAME 0 Key1 newName"
	parts, s, dbIndex := prepareCmdTest(cmd)

	s.Set(dbIndex, "Key1", "value1")

	aofRename(parts, s, dbIndex)
	value, ok := s.Get(dbIndex, "newName")
	valStr := value.Data.(string)

	if !ok || valStr != "value1" {
		t.Fatalf("Expeted 'value1, got %s", valStr)
	}
}

// Test aofLTrim
func TestAofLTrim(t *testing.T) {
	cmd := "LTRIM 0 List1 1 2"
	parts, s, dbIndex := prepareCmdTest(cmd)

	s.LPush(dbIndex, "List1", "Value1", "Value2", "Value3")
	aofLTrim(parts, s, dbIndex)
	list, _ := s.LRange(dbIndex, "List1", 0, -1)
	expectedList := []string{"Value2", "Value1"}
	listStr := make([]string, len(list))
	for i, v := range list {
		listStr[i] = v.(string)
	}
	if !slice.Equal(listStr, expectedList) {
		t.Logf("Expected %v, got %v", expectedList, list)
		t.Fail()
	}
}

func TestAOFWriterPersistsRESPCommandsLosslessly(t *testing.T) {
	filename := "resp_appendonly.aof"
	_ = os.Remove(filename)
	defer os.Remove(filename)

	aofChan := make(chan store.AOFCommand, 10)
	go AOFWriter(aofChan, filename)

	aofChan <- store.NewAOFCommand("SET", "0", "greeting", "hello world")
	aofChan <- store.NewAOFCommand("RPUSH", "0", "names", "Andre Cunha", "Maria Clara")
	aofChan <- store.NewAOFCommand("HSET", "0", "profile", "bio", "{\"full name\":\"Andre Cunha\"}")
	close(aofChan)

	time.Sleep(200 * time.Millisecond)

	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("failed to open AOF file: %v", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	proto := &resp2.RESP2Protocol{}

	assertCommand := func(expected []string) {
		t.Helper()
		value, err := proto.Parse(reader)
		if err != nil {
			t.Fatalf("failed to parse RESP command: %v", err)
		}
		parts, ok := respArrayToCommand(value)
		if !ok {
			t.Fatalf("failed to decode RESP command into parts")
		}
		if len(parts) != len(expected) {
			t.Fatalf("expected command %v, got %v", expected, parts)
		}
		for i := range expected {
			if parts[i] != expected[i] {
				t.Fatalf("expected command %v, got %v", expected, parts)
			}
		}
	}

	assertCommand([]string{"SET", "0", "greeting", "hello world"})
	assertCommand([]string{"RPUSH", "0", "names", "Andre Cunha", "Maria Clara"})
	assertCommand([]string{"HSET", "0", "profile", "bio", "{\"full name\":\"Andre Cunha\"}"})
}

func TestRebuildStoreFromLegacyAOFStartsEmpty(t *testing.T) {
	filename := "legacy_appendonly.aof"
	defer os.Remove(filename)

	if err := os.WriteFile(filename, []byte("SET 0 key hello world\n"), 0o666); err != nil {
		t.Fatalf("failed to write legacy AOF: %v", err)
	}

	s := store.NewStore(make(chan store.AOFCommand, 10))
	if err := RebuildStoreFromAOF(s, filename); err != nil {
		t.Fatalf("expected legacy AOF to be ignored, got error %v", err)
	}

	if s.Exists(0, "key") != 0 {
		t.Fatalf("expected store to remain empty when legacy AOF is ignored")
	}
}

func TestRebuildStoreFromAOFPreservesValuesWithSpaces(t *testing.T) {
	filename := "spaces_appendonly.aof"
	_ = os.Remove(filename)
	defer os.Remove(filename)

	aofChan := make(chan store.AOFCommand, 100)
	go AOFWriter(aofChan, filename)

	s := store.NewStore(aofChan)
	dbIndex := 0

	if _, err := s.Set(dbIndex, "title", "hello world"); err != nil {
		t.Fatalf("SET failed: %v", err)
	}
	s.RPush(dbIndex, "names", "Andre Cunha", "Maria Clara")
	if _, err := s.HSet(dbIndex, "profile", map[string]any{
		"display_name": "Andre Cunha",
		"bio":          "{\"summary\":\"hello world\"}",
	}); err != nil {
		t.Fatalf("HSET failed: %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	newStore := store.NewStore(make(chan store.AOFCommand, 100))
	if err := RebuildStoreFromAOF(newStore, filename); err != nil {
		t.Fatalf("Failed to rebuild state from AOF: %v", err)
	}

	value, ok := newStore.Get(dbIndex, "title")
	if !ok || value.Data.(string) != "hello world" {
		t.Fatalf("expected SET value with spaces to survive replay")
	}

	list, err := newStore.LRange(dbIndex, "names", 0, -1)
	if err != nil {
		t.Fatalf("LRANGE failed after replay: %v", err)
	}
	if len(list) != 2 || list[0] != "Andre Cunha" || list[1] != "Maria Clara" {
		t.Fatalf("expected list values with spaces to survive replay, got %v", list)
	}

	hashValue, ok, err := newStore.HGet(dbIndex, "profile", "display_name")
	if err != nil || !ok || hashValue != "Andre Cunha" {
		t.Fatalf("expected HSET value with spaces to survive replay, got ok=%v value=%q err=%v", ok, hashValue, err)
	}
}

func TestDispatchAOFReplaysAllPersistedMutations(t *testing.T) {
	s := store.NewStore(make(chan store.AOFCommand, 100))

	commands := [][]string{
		{"SET", "0", "counter", "1"},
		{"INCR", "0", "counter"},
		{"DECR", "0", "counter"},
		{"SETNX", "0", "unique", "first"},
		{"LPUSH", "0", "letters", "a", "b"},
		{"RPUSH", "0", "letters", "c"},
		{"LPOP", "0", "letters", "1"},
		{"RPOP", "0", "letters", "1"},
		{"HSET", "0", "profile", "name", "Andre Cunha", "city", "Sao Paulo"},
		{"HDEL", "0", "profile", "city"},
		{"SET", "0", "rename:old", "value"},
		{"RENAME", "0", "rename:old", "rename:new"},
		{"SET", "0", "flushdb:key", "value"},
		{"FLUSHDB", "0"},
		{"SET", "1", "db1:key", "value"},
		{"FLUSHALL"},
	}

	for _, cmd := range commands {
		dispatchAOF(cmd, s)
	}

	if s.Exists(0, "counter") != 0 {
		t.Fatalf("expected FLUSHALL to remove counter")
	}
	if s.Exists(1, "db1:key") != 0 {
		t.Fatalf("expected FLUSHALL to remove db1:key")
	}
}

func TestRespArrayToCommandRejectsNonArray(t *testing.T) {
	if _, ok := respArrayToCommand(protocol.BulkString([]byte("not-array"))); ok {
		t.Fatalf("expected non-array RESP value to be rejected")
	}
}

func prepareCmdTest(cmd string) ([]string, *store.Store, int) {
	aofChan := make(chan store.AOFCommand, 100)
	s := store.NewStore(aofChan)

	parts := cmdToParts(cmd)
	dbIndex, _ := strconv.Atoi(parts[1])
	return parts, s, dbIndex
}

func cmdToParts(cmd string) []string {
	switch cmd {
	case "RENAME 0 Key1 newName":
		return []string{"RENAME", "0", "Key1", "newName"}
	case "LTRIM 0 List1 1 2":
		return []string{"LTRIM", "0", "List1", "1", "2"}
	default:
		return []string{}
	}
}
