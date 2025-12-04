package store

// Delele a key from the store
func (s *Store) Delete(dbIndex int, key string) {
	s.delKey(dbIndex, key)
}

// delKey deletes a key from the store and its expiration (protected)
func (s *Store) delKey(dbIndex int, key string) {
	// mu
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data[dbIndex], key)
}

// flushDb flushes the database
func (s *Store) flushDb(dbIndex int) {
	s.data[dbIndex] = make(map[string]*Value)
}
