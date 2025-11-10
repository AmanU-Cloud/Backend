package user

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Caritas-Team/reviewer/internal/memcached"
	"github.com/bradfitz/gomemcache/memcache"
)

var (
	ErrRateLimitExceeded = errors.New("too many requests per minute")
	ErrInvalidConfig     = errors.New("invalid rate limiter configuration")
)

type RateLimiterInterface interface {
	AllowRequest(ctx context.Context, identifier string) (*RateLimitResult, error)
	ResetLimit(ctx context.Context, identifier string) error
}

type RateLimitResult struct {
	Allowed    bool
	Remaining  int
	RetryAfter time.Duration
	Error      error
}

type RateLimiter struct {
	cache             memcached.CacheInterface
	requestsPerMinute int
	bucketSize        int
	storageType       string
	enabled           bool
}

// Убедитесь, что RateLimiter реализует интерфейс
var _ RateLimiterInterface = (*RateLimiter)(nil)

func NewRateLimiter(cache memcached.CacheInterface, requestsPerMinute, bucketSize int, storageType string, enabled bool) *RateLimiter {
	return &RateLimiter{
		cache:             cache,
		requestsPerMinute: requestsPerMinute,
		bucketSize:        bucketSize,
		storageType:       storageType,
		enabled:           enabled,
	}
}

// AllowRequest проверяет, можно ли выполнить запрос по token bucket алгоритму
func (rl *RateLimiter) AllowRequest(ctx context.Context, identifier string) (*RateLimitResult, error) {
	if !rl.enabled {
		return &RateLimitResult{
			Allowed:   true,
			Remaining: rl.bucketSize,
		}, nil
	}

	switch rl.storageType {
	case "memcached", "memory":
		return rl.allowRequestTokenBucket(ctx, identifier)
	default:
		return nil, ErrInvalidConfig
	}
}

// Token Bucket алгоритм
func (rl *RateLimiter) allowRequestTokenBucket(ctx context.Context, identifier string) (*RateLimitResult, error) {
	key := "rate_limit:" + identifier
	now := time.Now().Unix()

	// Получаем текущее состояние bucket из кэша
	data, err := rl.cache.Get(ctx, key)

	var tokens int
	var lastRefill int64

	if err != nil {
		// Если ключа нет или ошибка - инициализируем новый bucket
		if err == memcache.ErrCacheMiss {
			tokens = rl.bucketSize
			lastRefill = now
		} else {
			// При других ошибках разрешаем запрос (fail open)
			return &RateLimitResult{
				Allowed:   true,
				Remaining: rl.bucketSize,
				Error:     fmt.Errorf("cache error: %w", err),
			}, nil
		}
	} else {
		// Парсим существующие данные из кэша
		parts := strings.Split(string(data), ":")
		if len(parts) != 2 {
			// Если данные повреждены, сбрасываем bucket
			tokens = rl.bucketSize - 1
			lastRefill = now
		} else {
			tokens, _ = strconv.Atoi(parts[0])
			lastRefill, _ = strconv.ParseInt(parts[1], 10, 64)
		}
	}

	// Вычисляем пополнение токенов
	tokensRefilled := rl.calculateRefill(now, lastRefill, tokens)
	tokens = tokensRefilled.tokens
	lastRefill = tokensRefilled.lastRefill

	// Проверяем достаточно ли токенов
	if tokens <= 0 {
		retryAfter := time.Duration(tokensRefilled.refillInterval) * time.Second
		return &RateLimitResult{
			Allowed:    false,
			Remaining:  0,
			RetryAfter: retryAfter,
			Error:      ErrRateLimitExceeded,
		}, nil
	}

	// Уменьшаем токены и сохраняем
	tokens--

	value := fmt.Sprintf("%d:%d", tokens, lastRefill)
	ttl := time.Duration(60*2) * time.Second // 2 минуты

	err = rl.cache.Set(ctx, key, []byte(value), ttl)
	if err != nil {
		// При ошибке сохранения разрешаем запрос (fail open)
		return &RateLimitResult{
			Allowed:   true,
			Remaining: tokens,
			Error:     fmt.Errorf("cache set error: %w", err),
		}, nil
	}

	return &RateLimitResult{
		Allowed:   true,
		Remaining: tokens,
	}, nil
}

type refillResult struct {
	tokens         int
	lastRefill     int64
	refillInterval int64
}

func (rl *RateLimiter) calculateRefill(now, lastRefill int64, currentTokens int) refillResult {
	// Интервал между пополнениями (в секундах)
	refillInterval := int64(60) / int64(rl.requestsPerMinute)

	if refillInterval <= 0 {
		refillInterval = 1
	}

	elapsed := now - lastRefill

	if elapsed <= 0 {
		return refillResult{
			tokens:         currentTokens,
			lastRefill:     lastRefill,
			refillInterval: refillInterval,
		}
	}

	// Вычисляем сколько токенов нужно пополнить
	refillTokens := int(elapsed / refillInterval)

	if refillTokens > 0 {
		newTokens := min(rl.bucketSize, currentTokens+refillTokens)
		return refillResult{
			tokens:         newTokens,
			lastRefill:     now,
			refillInterval: refillInterval,
		}
	}

	return refillResult{
		tokens:         currentTokens,
		lastRefill:     lastRefill,
		refillInterval: refillInterval,
	}
}

func (rl *RateLimiter) ResetLimit(ctx context.Context, identifier string) error {
	key := "rate_limit:" + identifier
	return rl.cache.Set(ctx, key, []byte("0:0"), time.Second)
}

func (rl *RateLimiter) GetRemainingTokens(ctx context.Context, identifier string) (int, error) {
	key := "rate_limit:" + identifier
	data, err := rl.cache.Get(ctx, key)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return rl.bucketSize, nil
		}
		return 0, err
	}

	parts := strings.Split(string(data), ":")
	if len(parts) != 2 {
		return rl.bucketSize, nil
	}

	tokens, _ := strconv.Atoi(parts[0])
	return tokens, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
