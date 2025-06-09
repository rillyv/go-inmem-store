package store

import (
	"sync"
	"time"
)

type TTLStore struct {
	mu sync.RWMutex
	ttlStore map[string]time.Time
	inMemoryStore *InMemoryStore
}

 func (t *TTLStore) Get(key string) (time.Time, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	val, ok := t.ttlStore[key]

	return val, ok
}

func (t *TTLStore) Expire(key string, ttlInSeconds int) string {
	t.mu.Lock()
	t.ttlStore[key] = time.Now().Add(time.Duration(ttlInSeconds) * time.Second)
	t.mu.Unlock()

	return "OK"
}

func (t *TTLStore) GetTtl(key string) int {
	if t.inMemoryStore.evictWhenExpired(key) {
		return -2
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.inMemoryStore.KeyExists(key) {
		return -2
	}

	ttl, ok := t.ttlStore[key]
	if !ok {
		return -1
	}

	return int(ttl.Unix() - time.Now().Unix())
}

func (t *TTLStore) Delete(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.ttlStore, key)
}
