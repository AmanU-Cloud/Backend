package user

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/Caritas-Team/reviewer/internal/config"
	"github.com/Caritas-Team/reviewer/internal/memcached"
)

var ErrRateLimitExceeded = errors.New("too many requests")

type RateLimiter struct {
	cache    memcached.CacheInterface
	enabled  bool
	window   time.Duration
	requests int
}

func NewRateLimiter(cache memcached.CacheInterface, cfg config.Config) *RateLimiter {
	return &RateLimiter{
		cache:    cache,
		enabled:  cfg.RateLimiter.Enabled,
		window:   time.Duration(cfg.RateLimiter.WindowSize) * time.Second,
		requests: cfg.RateLimiter.RequestsPerWindow,
	}
}

func (rl *RateLimiter) AllowRequest(ctx context.Context, userID string) error {
	if !rl.enabled {
		return nil
	}

	key := "rate_limit:" + userID

	_, err := rl.cache.Get(ctx, key)
	if err == nil {
		return ErrRateLimitExceeded
	}

	err = rl.cache.Set(ctx, key, []byte("1"), rl.window)
	if err != nil {
		slog.Info("Rate limit check", "error", err, "key", key)
		return nil
	}

	return nil
}
