package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type APIKeyStatus string

const (
	APIKeyStatusActive  APIKeyStatus = "active"
	APIKeyStatusRevoked APIKeyStatus = "revoked"
)

type APIKey struct {
	ID             string
	UserID         string
	KeyPrefix      string
	KeyHash        string
	Name           string
	Status         APIKeyStatus
	RateLimitRPS   int
	RateLimitBurst int
	CreatedAt      time.Time
	ExpiresAt      *time.Time
}

// Store is the abstraction for API key lookup (and later: usage logging, issuance).
type Store interface {
	LookupAPIKey(ctx context.Context, rawKey string) (*APIKey, error)
	// GetProviderCredential returns the encrypted provider API key for the given user/provider/name.
	// Returns ("" , error) when not found.
	GetProviderCredential(ctx context.Context, userID, provider, name string) (string, error)
	// ListAPIKeys returns up to `limit` API keys for inspection/logging.
	ListAPIKeys(ctx context.Context, limit int) ([]APIKey, error)
	Close(ctx context.Context) error
}

// HashAPIKey returns a hex-encoded SHA-256 of the raw key.
// We intentionally hash the entire raw key string.
func HashAPIKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
