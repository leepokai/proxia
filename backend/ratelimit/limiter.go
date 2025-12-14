package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// KeyedLimiter keeps a rate limiter per key ID.
type KeyedLimiter struct {
	mu  sync.Mutex
	lim map[string]*rate.Limiter
}

func NewKeyedLimiter() *KeyedLimiter {
	return &KeyedLimiter{
		lim: make(map[string]*rate.Limiter),
	}
}

func (k *KeyedLimiter) Allow(keyID string, rps, burst int) bool {
	k.mu.Lock()
	defer k.mu.Unlock()
	lim, ok := k.lim[keyID]
	if !ok {
		lim = rate.NewLimiter(rate.Limit(rps), burst)
		k.lim[keyID] = lim
	} else {
		// Adjust if the config has changed materially
		lim.SetLimit(rate.Limit(rps))
		lim.SetBurst(burst)
	}
	return lim.Allow()
}

// ReserveWait waits up to maxWait for a token. Not used in MVP.
func (k *KeyedLimiter) ReserveWait(keyID string, rps, burst int, maxWait time.Duration) bool {
	k.mu.Lock()
	lim, ok := k.lim[keyID]
	if !ok {
		lim = rate.NewLimiter(rate.Limit(rps), burst)
		k.lim[keyID] = lim
	} else {
		lim.SetLimit(rate.Limit(rps))
		lim.SetBurst(burst)
	}
	k.mu.Unlock()
	r := lim.Reserve()
	if !r.OK() {
		return false
	}
	if r.Delay() > maxWait {
		r.Cancel()
		return false
	}
	time.Sleep(r.Delay())
	return true
}
