package rdb

import (
	"os"
	"testing"
	"time"

	"github.com/ottermq/goodiesdb/internal/core/store"
	"github.com/ottermq/goodiesdb/internal/persistence/aof"
)

func TestSaveLoadSnapshot(t *testing.T) {
	// Create a temporary AOF file
	aofFilename := "test_appendonly.aof"
	aofChan := make(chan store.AOFCommand, 100)
	dbIndex := 0

	// Start the AOF writer
	go aof.AOFWriter(aofChan, aofFilename)

	// Initialize a new store with the AOF file
	s := store.NewStore(aofChan)

	s.Set(dbIndex, "Key1", "Value1")
	s.Set(dbIndex, "Key2", "Value2")
	s.Expire(dbIndex, "Key1", 3*time.Second)

	err := SaveSnapshot(s, "test_snapshot.gob")
	if err != nil {
		t.Fatalf("Failed to save snapshot: %v", err)
	}

	newStore := store.NewStore(aofChan)
	err = LoadSnapshot(newStore, "test_snapshot.gob")
	if err != nil {
		t.Fatalf("Failed to load snapshot: %v", err)
	}

	// Verify Key1 exists before it expires
	value, ok := newStore.Get(dbIndex, "Key1")
	valStr := value.Data.(string)
	if !ok || valStr != "Value1" {
		t.Fatalf("Expected Value1, got %s", valStr)
	}

	// Verify Key2 exists before it expires
	value, ok = newStore.Get(dbIndex, "Key2")
	valStr = value.Data.(string)
	if !ok || valStr != "Value2" {
		t.Fatalf("Expected Value2, got %s", valStr)
	}

	// Wait for the key to expire
	time.Sleep(4 * time.Second)

	// Verify Key1 exists after it expires
	if newStore.Exists(dbIndex, "Key1") > 0 {
		t.Fatalf("Expected Key1 to be expered after snapshot load an waiting more than 3 seconds")
	}

	// Clean up the snapshot file
	err = os.Remove("test_snapshot.gob")

	// Clean up the AOF file
	os.Remove(aofFilename)

}

func TestSaveLoadSnapshotWithHash(t *testing.T) {
	aofFilename := "test_hash_appendonly.aof"
	aofChan := make(chan store.AOFCommand, 100)

	go aof.AOFWriter(aofChan, aofFilename)

	s := store.NewStore(aofChan)
	if _, err := s.HSet(0, "profile", map[string]any{
		"name": "andre",
		"role": "admin",
	}); err != nil {
		t.Fatalf("HSet failed: %v", err)
	}

	if err := SaveSnapshot(s, "test_hash_snapshot.gob"); err != nil {
		t.Fatalf("Failed to save snapshot with hash: %v", err)
	}

	newStore := store.NewStore(aofChan)
	if err := LoadSnapshot(newStore, "test_hash_snapshot.gob"); err != nil {
		t.Fatalf("Failed to load snapshot with hash: %v", err)
	}

	hash, err := newStore.HGetAll(0, "profile")
	if err != nil {
		t.Fatalf("HGetAll failed after load: %v", err)
	}
	if hash["name"] != "andre" || hash["role"] != "admin" {
		t.Fatalf("unexpected hash after load: %#v", hash)
	}

	if err := os.Remove("test_hash_snapshot.gob"); err != nil {
		t.Fatalf("failed to remove hash snapshot: %v", err)
	}
	_ = os.Remove(aofFilename)
}
