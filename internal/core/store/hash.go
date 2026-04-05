package store

import (
	"fmt"
	"sort"
)

func (s *Store) HSet(dbIndex int, key string, fields map[string]any) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	value, ok := s.data[dbIndex][key]
	if ok {
		if value.IsExpired() {
			s.delKey(dbIndex, key)
			ok = false
		} else if value.Type != TypeHash {
			return 0, ErrWrongType
		}
	}

	var hash map[string]any
	if !ok {
		hash = make(map[string]any)
		value = NewHashValue(hash)
		s.data[dbIndex][key] = value
	} else {
		var err error
		hash, err = value.AsHash()
		if err != nil {
			return 0, err
		}
	}

	added := 0
	for field, rawValue := range fields {
		if _, exists := hash[field]; !exists {
			added++
		}
		hash[field] = fmt.Sprintf("%v", rawValue)
	}

	s.appendAOF("HSET", append([]string{dbIndexArg(dbIndex), key}, flattenHashFieldPairs(fields)...)...)
	return added, nil
}

func (s *Store) HGet(dbIndex int, key, field string) (string, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.data[dbIndex][key]
	if !ok || value.IsExpired() {
		return "", false, nil
	}
	hash, err := value.AsHash()
	if err != nil {
		return "", false, err
	}
	rawValue, exists := hash[field]
	if !exists {
		return "", false, nil
	}
	return fmt.Sprintf("%v", rawValue), true, nil
}

func (s *Store) HGetAll(dbIndex int, key string) (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.data[dbIndex][key]
	if !ok || value.IsExpired() {
		return map[string]string{}, nil
	}
	hash, err := value.AsHash()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(hash))
	for field, rawValue := range hash {
		result[field] = fmt.Sprintf("%v", rawValue)
	}
	return result, nil
}

func (s *Store) HDel(dbIndex int, key string, fields ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	value, ok := s.data[dbIndex][key]
	if !ok {
		return 0, nil
	}
	if value.IsExpired() {
		s.delKey(dbIndex, key)
		return 0, nil
	}

	hash, err := value.AsHash()
	if err != nil {
		return 0, err
	}

	deleted := 0
	for _, field := range fields {
		if _, exists := hash[field]; exists {
			delete(hash, field)
			deleted++
		}
	}

	if len(hash) == 0 {
		s.delKey(dbIndex, key)
	}

	if deleted > 0 {
		s.appendAOF("HDEL", append([]string{dbIndexArg(dbIndex), key}, fields...)...)
	}

	return deleted, nil
}

func (s *Store) HExists(dbIndex int, key, field string) (bool, error) {
	_, ok, err := s.HGet(dbIndex, key, field)
	return ok, err
}

func (s *Store) HLen(dbIndex int, key string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.data[dbIndex][key]
	if !ok || value.IsExpired() {
		return 0, nil
	}
	hash, err := value.AsHash()
	if err != nil {
		return 0, err
	}
	return len(hash), nil
}

func (s *Store) HMGet(dbIndex int, key string, fields ...string) ([]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]any, len(fields))

	value, ok := s.data[dbIndex][key]
	if !ok || value.IsExpired() {
		return result, nil
	}
	hash, err := value.AsHash()
	if err != nil {
		return nil, err
	}

	for i, field := range fields {
		if rawValue, exists := hash[field]; exists {
			result[i] = fmt.Sprintf("%v", rawValue)
		}
	}
	return result, nil
}

func (s *Store) HKeys(dbIndex int, key string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.data[dbIndex][key]
	if !ok || value.IsExpired() {
		return []string{}, nil
	}
	hash, err := value.AsHash()
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(hash))
	for field := range hash {
		keys = append(keys, field)
	}
	return keys, nil
}

func (s *Store) HVals(dbIndex int, key string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.data[dbIndex][key]
	if !ok || value.IsExpired() {
		return []string{}, nil
	}
	hash, err := value.AsHash()
	if err != nil {
		return nil, err
	}

	values := make([]string, 0, len(hash))
	for _, rawValue := range hash {
		values = append(values, fmt.Sprintf("%v", rawValue))
	}
	return values, nil
}

func flattenHashFieldPairs(fields map[string]any) []string {
	keys := make([]string, 0, len(fields))
	for field := range fields {
		keys = append(keys, field)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(fields)*2)
	for _, field := range keys {
		parts = append(parts, field, fmt.Sprintf("%v", fields[field]))
	}
	return parts
}
