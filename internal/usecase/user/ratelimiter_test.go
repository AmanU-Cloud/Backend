package user

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

// MockCache для тестирования
type MockCache struct {
	GetFunc   func(ctx context.Context, key string) ([]byte, error)
	SetFunc   func(ctx context.Context, key string, value []byte, ttl time.Duration) error
	CloseFunc func() error
}

func (m *MockCache) Get(ctx context.Context, key string) ([]byte, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key)
	}
	return nil, memcache.ErrCacheMiss
}

func (m *MockCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if m.SetFunc != nil {
		return m.SetFunc(ctx, key, value, ttl)
	}
	return nil
}

func (m *MockCache) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func TestNewRateLimiter(t *testing.T) {
	mockCache := &MockCache{}
	rl := NewRateLimiter(mockCache, 60, 60, "memcached", true)

	if rl.requestsPerMinute != 60 {
		t.Errorf("Expected 60 requests per minute, got %d", rl.requestsPerMinute)
	}
	if rl.bucketSize != 60 {
		t.Errorf("Expected 60 bucket size, got %d", rl.bucketSize)
	}
	if rl.storageType != "memcached" {
		t.Errorf("Expected storage type 'memcached', got %s", rl.storageType)
	}
	if !rl.enabled {
		t.Error("Expected rate limiter to be enabled")
	}
}

func TestRateLimiter_AllowRequest_Disabled(t *testing.T) {
	mockCache := &MockCache{}
	rl := NewRateLimiter(mockCache, 60, 60, "memcached", false)

	result, err := rl.AllowRequest(context.Background(), "test-ip")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Error("Expected request to be allowed when rate limiter is disabled")
	}
}

func TestRateLimiter_AllowRequest_FirstRequest(t *testing.T) {
	mockCache := &MockCache{
		GetFunc: func(ctx context.Context, key string) ([]byte, error) {
			return nil, memcache.ErrCacheMiss
		},
		SetFunc: func(ctx context.Context, key string, value []byte, ttl time.Duration) error {
			return nil
		},
	}

	rl := NewRateLimiter(mockCache, 10, 10, "memcached", true)

	result, err := rl.AllowRequest(context.Background(), "test-ip")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Error("Expected first request to be allowed")
	}
	if result.Remaining != 9 {
		t.Errorf("Expected 9 remaining tokens, got %d", result.Remaining)
	}
}

func TestRateLimiter_AllowRequest_RateLimitExceeded(t *testing.T) {
	callCount := 0
	now := time.Now().Unix()

	mockCache := &MockCache{
		GetFunc: func(ctx context.Context, key string) ([]byte, error) {
			if callCount == 0 {
				callCount++
				return nil, memcache.ErrCacheMiss
			}
			// После первого запроса возвращаем состояние с 0 токенов и ТЕКУЩИМ временем
			// чтобы не было пополнения токенов
			return []byte("0:" + strconv.FormatInt(now, 10)), nil
		},
		SetFunc: func(ctx context.Context, key string, value []byte, ttl time.Duration) error {
			return nil
		},
	}

	rl := NewRateLimiter(mockCache, 10, 10, "memcached", true)

	// Первый запрос - должен пройти
	result, err := rl.AllowRequest(context.Background(), "test-ip")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Error("Expected first request to be allowed")
	}

	// Второй запрос - должен быть отклонен
	result, err = rl.AllowRequest(context.Background(), "test-ip")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Allowed {
		t.Error("Expected second request to be blocked due to rate limit")
	}
	if result.Error == nil || !errors.Is(result.Error, ErrRateLimitExceeded) {
		t.Error("Expected rate limit exceeded error")
	}
}

func TestRateLimiter_AllowRequest_CacheError(t *testing.T) {
	mockCache := &MockCache{
		GetFunc: func(ctx context.Context, key string) ([]byte, error) {
			return nil, errors.New("cache connection error")
		},
		SetFunc: func(ctx context.Context, key string, value []byte, ttl time.Duration) error {
			return nil
		},
	}

	rl := NewRateLimiter(mockCache, 10, 10, "memcached", true)

	// При ошибке кэша запрос должен быть разрешен (fail open)
	result, err := rl.AllowRequest(context.Background(), "test-ip")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Error("Expected request to be allowed on cache error")
	}
	if result.Error == nil {
		t.Error("Expected error in result")
	}
}

func TestRateLimiter_AllowRequest_TokenRefill(t *testing.T) {
	now := time.Now().Unix()
	// Устанавливаем lastRefill на 10 секунд назад (больше чем интервал пополнения)
	oldLastRefill := now - 10

	mockCache := &MockCache{
		GetFunc: func(ctx context.Context, key string) ([]byte, error) {
			// Возвращаем состояние с 0 токенов, но старым временем
			return []byte("0:" + strconv.FormatInt(oldLastRefill, 10)), nil
		},
		SetFunc: func(ctx context.Context, key string, value []byte, ttl time.Duration) error {
			return nil
		},
	}

	rl := NewRateLimiter(mockCache, 60, 10, "memcached", true)
	// Интервал пополнения = 1 секунда (60/60)
	// За 10 секунд должно пополниться 10 токенов

	result, err := rl.AllowRequest(context.Background(), "test-ip")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Error("Expected request to be allowed after token refill")
	}
}

func TestRateLimiter_AllowRequest_CorruptedData(t *testing.T) {
	mockCache := &MockCache{
		GetFunc: func(ctx context.Context, key string) ([]byte, error) {
			// Возвращаем поврежденные данные
			return []byte("invalid:data:format"), nil
		},
		SetFunc: func(ctx context.Context, key string, value []byte, ttl time.Duration) error {
			return nil
		},
	}

	rl := NewRateLimiter(mockCache, 10, 10, "memcached", true)

	// При поврежденных данных должен создаться новый bucket
	result, err := rl.AllowRequest(context.Background(), "test-ip")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Error("Expected request to be allowed with corrupted data")
	}
}

func TestRateLimiter_InvalidStorageType(t *testing.T) {
	mockCache := &MockCache{}
	rl := NewRateLimiter(mockCache, 60, 60, "invalid_storage", true)

	result, err := rl.AllowRequest(context.Background(), "test-ip")
	if err == nil {
		t.Error("Expected error for invalid storage type")
	}
	if result != nil {
		t.Error("Expected nil result for invalid storage type")
	}
}

func TestRateLimiter_ResetLimit(t *testing.T) {
	setCalled := false
	mockCache := &MockCache{
		SetFunc: func(ctx context.Context, key string, value []byte, ttl time.Duration) error {
			setCalled = true
			return nil
		},
	}

	rl := NewRateLimiter(mockCache, 60, 60, "memcached", true)
	err := rl.ResetLimit(context.Background(), "test-ip")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !setCalled {
		t.Error("Expected Set to be called for reset")
	}
}

func TestRateLimiter_GetRemainingTokens(t *testing.T) {
	mockCache := &MockCache{
		GetFunc: func(ctx context.Context, key string) ([]byte, error) {
			return []byte("5:1234567890"), nil
		},
	}

	rl := NewRateLimiter(mockCache, 10, 10, "memcached", true)
	tokens, err := rl.GetRemainingTokens(context.Background(), "test-ip")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if tokens != 5 {
		t.Errorf("Expected 5 remaining tokens, got %d", tokens)
	}
}

func TestRateLimiter_GetRemainingTokens_CacheMiss(t *testing.T) {
	mockCache := &MockCache{
		GetFunc: func(ctx context.Context, key string) ([]byte, error) {
			return nil, memcache.ErrCacheMiss
		},
	}

	rl := NewRateLimiter(mockCache, 10, 10, "memcached", true)
	tokens, err := rl.GetRemainingTokens(context.Background(), "test-ip")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if tokens != 10 {
		t.Errorf("Expected 10 remaining tokens (bucket size) on cache miss, got %d", tokens)
	}
}

func TestMinFunction(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{0, 5, 0},
		{5, 0, 0},
		{-1, 1, -1},
	}

	for _, test := range tests {
		result := min(test.a, test.b)
		if result != test.expected {
			t.Errorf("min(%d, %d) = %d, expected %d", test.a, test.b, result, test.expected)
		}
	}
}
