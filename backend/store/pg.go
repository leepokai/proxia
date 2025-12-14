package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PGStore struct {
	pool *pgxpool.Pool
}

func NewPGStore(ctx context.Context, dsn string) (*PGStore, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect db: %w", err)
	}
	return &PGStore{pool: pool}, nil
}

func (s *PGStore) Close(ctx context.Context) error {
	if s.pool != nil {
		s.pool.Close()
	}
	return nil
}

func (s *PGStore) LookupAPIKey(ctx context.Context, rawKey string) (*APIKey, error) {
	hash := HashAPIKey(rawKey)
	const q = `
select id, user_id, key_prefix, key_hash, coalesce(name,'') as name, status, 
       rate_limit_rps, rate_limit_burst, created_at, expires_at
from public.api_keys
where key_hash = $1
  and status = 'active'
  and (expires_at is null or expires_at > now())
limit 1`
	row := s.pool.QueryRow(ctx, q, hash)
	var k APIKey
	var expiresAt *time.Time
	if err := row.Scan(
		&k.ID, &k.UserID, &k.KeyPrefix, &k.KeyHash, &k.Name, &k.Status,
		&k.RateLimitRPS, &k.RateLimitBurst, &k.CreatedAt, &expiresAt,
	); err != nil {
		return nil, errors.New("invalid api key")
	}
	k.ExpiresAt = expiresAt
	return &k, nil
}

func (s *PGStore) GetProviderCredential(ctx context.Context, userID, provider, name string) (string, error) {
	if name == "" {
		name = "default"
	}
	const q = `
select encrypted_key
from public.provider_credentials
where user_id = $1 and provider = $2 and name = $3
limit 1`
	var enc string
	if err := s.pool.QueryRow(ctx, q, userID, provider, name).Scan(&enc); err != nil {
		return "", errors.New("provider credential not found")
	}
	return enc, nil
}

func (s *PGStore) ListAPIKeys(ctx context.Context, limit int) ([]APIKey, error) {
	if limit <= 0 {
		limit = 20
	}
	const q = `
select id, user_id, key_prefix, key_hash, coalesce(name,'') as name, status,
       rate_limit_rps, rate_limit_burst, created_at, expires_at
from public.api_keys
order by created_at desc
limit $1`
	rows, err := s.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []APIKey
	for rows.Next() {
		var k APIKey
		var expiresAt *time.Time
		if err := rows.Scan(
			&k.ID, &k.UserID, &k.KeyPrefix, &k.KeyHash, &k.Name, &k.Status,
			&k.RateLimitRPS, &k.RateLimitBurst, &k.CreatedAt, &expiresAt,
		); err != nil {
			return nil, err
		}
		k.ExpiresAt = expiresAt
		out = append(out, k)
	}
	return out, rows.Err()
}
