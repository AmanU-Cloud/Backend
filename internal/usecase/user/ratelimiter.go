package user

import (
	"context"
	"errors"
	"time"

	"github.com/Caritas-Team/reviewer/internal/memcached"
	"golang.org/x/sys/windows"
)

var (
	ErrRateLimitExceeded = errors.New("too many requests per minute")
)

type RateLimiter struct {
	cache             memcached.CacheInterface
	requestsPerMinute int
	window            time.Duration
	enabled           bool
}

func NewRateLimiter(cache memcached.CacheInterface, requestsPerMinute int, window time.Duration, enabled bool) *RateLimiter {
	return &RateLimiter{
		cache:             cache,
		requestsPerMinute: requestsPerMinute,
		window:            window,
		enabled:           enabled,
	}
}

func (rl *RateLimiter) AllowRequest(ctx context.Context, identifier string) error {
	key := "rate_limit_" + identifier
	if !rl.enabled {
		return nil
	}
	get, err := rl.cache.Get(ctx, key)
	if get == nil {
		return nil
	}
	if err != nil {
		return nil
	}
	return nil
}
