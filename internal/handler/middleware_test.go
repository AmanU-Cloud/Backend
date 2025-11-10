package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Caritas-Team/reviewer/internal/config"
	"github.com/Caritas-Team/reviewer/internal/usecase/user"
)

// MockRateLimiter реализует user.RateLimiterInterface
type MockRateLimiter struct {
	AllowRequestFunc func(ctx context.Context, identifier string) (*user.RateLimitResult, error)
	ResetLimitFunc   func(ctx context.Context, identifier string) error
}

func (m *MockRateLimiter) AllowRequest(ctx context.Context, identifier string) (*user.RateLimitResult, error) {
	if m.AllowRequestFunc != nil {
		return m.AllowRequestFunc(ctx, identifier)
	}
	return &user.RateLimitResult{Allowed: true, Remaining: 10}, nil
}

func (m *MockRateLimiter) ResetLimit(ctx context.Context, identifier string) error {
	if m.ResetLimitFunc != nil {
		return m.ResetLimitFunc(ctx, identifier)
	}
	return nil
}

func TestRateLimiterMiddleware_OptionsRequest(t *testing.T) {
	mockLimiter := &MockRateLimiter{
		AllowRequestFunc: func(ctx context.Context, identifier string) (*user.RateLimitResult, error) {
			t.Error("Rate limiter should not be called for OPTIONS requests")
			return nil, nil
		},
	}

	cfg := &config.Config{}
	middleware := NewRateLimiterMiddleware(mockLimiter, cfg)

	handlerCalled := false
	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !handlerCalled {
		t.Error("Handler should be called for OPTIONS requests")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 for OPTIONS, got %d", rr.Code)
	}
}
