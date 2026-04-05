package store

import "strconv"

// Delele a key from the store
func (s *Store) Delete(dbIndex int, key string) {
	s.delKey(dbIndex, key)
}

func (s *Store) appendAOF(name string, args ...string) {
	if s.aofChan == nil {
		return
	}
	s.aofChan <- NewAOFCommand(name, args...)
}

func dbIndexArg(dbIndex int) string {
	return strconv.Itoa(dbIndex)
}

// delKey deletes a key from the store and its expiration (protected)
func (s *Store) delKey(dbIndex int, key string) {
	delete(s.data[dbIndex], key)
}

// flushDb flushes the database
func (s *Store) flushDb(dbIndex int) {
	s.data[dbIndex] = make(map[string]*Value)
}
