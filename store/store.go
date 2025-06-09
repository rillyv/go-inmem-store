package store

import (
	"sync"
	"time"
)

type InMemoryStore struct {
	mu   sync.RWMutex
	data map[string]string
	ttl  *TTLStore
}

func NewInMemoryStore() *InMemoryStore {
	inMemoryStore := &InMemoryStore{
		data: make(map[string]string),
	}

	ttlStore := &TTLStore{
		ttlStore: make(map[string]time.Time),
		inMemoryStore: inMemoryStore,
	}

	inMemoryStore.ttl = ttlStore

	return inMemoryStore
}

func (t *InMemoryStore) RunTTLCleanup() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for range ticker.C {
		expiredKeys := []string{}

		t.mu.RLock()
		for k, v := range t.ttl.ttlStore {
			if v.Unix() < time.Now().Unix() {
				expiredKeys = append(expiredKeys, k)
			}
		}
		t.mu.RUnlock()

		t.mu.Lock()
		for _, v := range expiredKeys {
			delete(t.data, v)
			delete(t.ttl.ttlStore, v)
		}
		t.mu.Unlock()
	}
}

func (s *InMemoryStore) KeyExists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.data[key]
	return ok
}

func (s *InMemoryStore) deleteKeyFromStoreAndTtl(key string) bool {
	if s.KeyExists(key) {
		s.mu.Lock()
		delete(s.data, key)
		s.mu.Unlock()

		s.ttl.Delete(key)

		return true
	} else {
		return false
	}
}

func (s *InMemoryStore) evictWhenExpired(key string) bool {
	s.mu.RLock()
	ttl, ok := s.ttl.Get(key)
	s.mu.RUnlock()

	if !ok {
		return false
	}

	if ttl.Unix() < time.Now().Unix() {
		return s.deleteKeyFromStoreAndTtl(key)
	}

	return false
}

func (s *InMemoryStore) Set(key string, value string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = value

	return "OK"
}

func (s *InMemoryStore) Get(key string) string {
	if s.evictWhenExpired(key) {
		return "NULL"
	}

	s.mu.RLock()
	val, ok := s.data[key]
	s.mu.RUnlock()

	if !ok {
		return "NULL"
	} else {
		return val
	}
}

func (s *InMemoryStore) Delete(key string) string {
	if s.evictWhenExpired(key) {
		return "NULL"
	}

	if s.deleteKeyFromStoreAndTtl(key) {
		return "OK"
	}

	return "NULL"
}

func (s *InMemoryStore) Expire(key string, ttlInSeconds int) string {
	if !s.KeyExists(key) {
		return "NULL"
	}

	return s.ttl.Expire(key, ttlInSeconds)
}

func (s *InMemoryStore) GetTtl(key string) int {
	return s.ttl.inMemoryStore.GetTtl(key)
}

func (s *InMemoryStore) Keys(offset int, limit int) []string {
	keys := []string{}

	s.mu.RLock()
	for k := range s.data {
		keys = append(keys, k)
	}
	s.mu.RUnlock()

	var end int
	if (offset + limit) > len(keys) {
		end = len(keys)
	} else {
		end = offset + limit
	}

	if offset > len(keys) {
		return []string{}
	}

	return keys[offset:end]
}

func (s *InMemoryStore) FlushAll() {
	s.mu.Lock()
	s.data = map[string]string{}
	s.ttl.ttlStore = map[string]time.Time{}
	s.mu.Unlock()
}
