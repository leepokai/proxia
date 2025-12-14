package auth

import (
	"context"
	"net/http"
	"strings"

	"goproject/ratelimit"
	"goproject/store"
	"goproject/utils"
)

type ctxKey string

const (
	apiKeyCtxKey      ctxKey = "gateway_api_key"
	upstreamKeyCtxKey ctxKey = "upstream_api_key"
)

func APIKeyFromContext(ctx context.Context) *store.APIKey {
	if v := ctx.Value(apiKeyCtxKey); v != nil {
		if k, ok := v.(*store.APIKey); ok {
			return k
		}
	}
	return nil
}

// UpstreamKeyFromContext returns a raw upstream provider key (when caller passed a non-gateway key).
func UpstreamKeyFromContext(ctx context.Context) string {
	if v := ctx.Value(upstreamKeyCtxKey); v != nil {
		if k, ok := v.(string); ok {
			return k
		}
	}
	return ""
}

// Middleware returns a wrapper that:
// - extracts Bearer token
// - validates it against the Store
// - enforces per-key rate limit
func Middleware(s store.Store, limiter *ratelimit.KeyedLimiter, logger *utils.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authz := r.Header.Get("Authorization")
			if authz == "" || !strings.HasPrefix(strings.ToLower(authz), "bearer ") {
				utils.WriteError(w, http.StatusUnauthorized, "missing bearer token")
				return
			}
			rawKey := strings.TrimSpace(authz[len("Bearer "):])
			if rawKey == "" {
				utils.WriteError(w, http.StatusUnauthorized, "invalid bearer token")
				return
			}
			if k, err := s.LookupAPIKey(r.Context(), rawKey); err == nil && k != nil {
				// Gateway key path with rate limiting.
				if !limiter.Allow(k.ID, k.RateLimitRPS, k.RateLimitBurst) {
					utils.WriteError(w, http.StatusTooManyRequests, "rate limit exceeded")
					return
				}
				ctx := context.WithValue(r.Context(), apiKeyCtxKey, k)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Fallback: treat as upstream provider key (no gateway lookup / rate limit).
			// Caller is responsible for passing provider/model so routing picks the right backend.
			ctx := context.WithValue(r.Context(), upstreamKeyCtxKey, rawKey)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
