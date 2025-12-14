package store

import (
	"context"
	"errors"
	"sync"
	"time"
)

// MemoryStore is a simple in-memory implementation for development.
type MemoryStore struct {
	mu  sync.RWMutex
	byH map[string]*APIKey // key_hash -> APIKey
}

func NewMemoryStore(devRawKey string) *MemoryStore {
	ms := &MemoryStore{
		byH: make(map[string]*APIKey),
	}
	if devRawKey != "" {
		now := time.Now()
		ms.byH[HashAPIKey(devRawKey)] = &APIKey{
			ID:             "dev-" + now.Format("20060102150405"),
			UserID:         "dev-user",
			KeyPrefix:      prefixFromRaw(devRawKey),
			KeyHash:        HashAPIKey(devRawKey),
			Name:           "Development Key",
			Status:         APIKeyStatusActive,
			RateLimitRPS:   3,
			RateLimitBurst: 10,
			CreatedAt:      now,
			ExpiresAt:      nil,
		}
	}
	return ms
}

func (m *MemoryStore) LookupAPIKey(ctx context.Context, rawKey string) (*APIKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	h := HashAPIKey(rawKey)
	k, ok := m.byH[h]
	if !ok {
		return nil, errors.New("invalid api key")
	}
	// Copy to avoid external modification
	cp := *k
	// Expiration check
	if cp.ExpiresAt != nil && time.Now().After(*cp.ExpiresAt) {
		return nil, errors.New("api key expired")
	}
	if cp.Status != APIKeyStatusActive {
		return nil, errors.New("api key not active")
	}
	return &cp, nil
}

func (m *MemoryStore) Close(ctx context.Context) error { return nil }

func (m *MemoryStore) GetProviderCredential(ctx context.Context, userID, provider, name string) (string, error) {
	return "", errors.New("provider credential not found")
}

func (m *MemoryStore) ListAPIKeys(ctx context.Context, limit int) ([]APIKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]APIKey, 0, len(m.byH))
	for _, k := range m.byH {
		out = append(out, *k)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

func prefixFromRaw(raw string) string {
	// best-effort prefix extraction:
	// return the first 6-8 chars or up to the first underscore
	end := len(raw)
	for i := 0; i < len(raw); i++ {
		if raw[i] == '_' {
			end = i + 1
			break
		}
	}
	if end > 12 {
		end = 12
	}
	return raw[:end]
}
