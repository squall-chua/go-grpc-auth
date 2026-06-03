package auth

import (
	"context"
	"sync"
	"time"
)

type memoryNonceEntry struct {
	expiresAt time.Time
}

type memoryNonceStore struct {
	mu   sync.Mutex
	data map[string]memoryNonceEntry
}

func NewMemoryNonceStore() NonceStore {
	return &memoryNonceStore{data: make(map[string]memoryNonceEntry)}
}

func nonceKey(ns, wallet, nonce string) string {
	return ns + "|" + wallet + "|" + nonce
}

func (s *memoryNonceStore) Save(_ context.Context, ns, wallet, nonce string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[nonceKey(ns, wallet, nonce)] = memoryNonceEntry{expiresAt: time.Now().UTC().Add(ttl)}
	return nil
}

func (s *memoryNonceStore) Consume(_ context.Context, ns, wallet, nonce string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := nonceKey(ns, wallet, nonce)
	entry, ok := s.data[k]
	if !ok {
		return false, nil
	}
	delete(s.data, k)
	if time.Now().UTC().After(entry.expiresAt) {
		return false, nil
	}
	return true, nil
}
