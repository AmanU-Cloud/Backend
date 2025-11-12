package handler

import (
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Caritas-Team/reviewer/internal/config"
	"github.com/Caritas-Team/reviewer/internal/usecase/user"
	"github.com/rs/cors"
)

var (
	DefaultCORSMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	DefaultCORSHeaders = []string{"Content-Type", "Authorization"}
)

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAgeSeconds    int
}

func CORS(cfg CORSConfig) func(http.Handler) http.Handler {
	if len(cfg.AllowedMethods) == 0 {
		cfg.AllowedMethods = DefaultCORSMethods
	}
	if len(cfg.AllowedHeaders) == 0 {
		cfg.AllowedHeaders = DefaultCORSHeaders
	}

	if len(cfg.AllowedOrigins) == 0 {
		c := cors.AllowAll()
		return func(next http.Handler) http.Handler { return c.Handler(next) }
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   cfg.AllowedMethods,
		AllowedHeaders:   cfg.AllowedHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAgeSeconds,
	})
	return func(next http.Handler) http.Handler { return c.Handler(next) }
}

type RateLimiterMiddleware struct {
	rateLimiter user.RateLimiter
	config      *config.Config
}

func NewRateLimiterMiddleware(rateLimiter user.RateLimiter, cfg *config.Config) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		rateLimiter: rateLimiter,
		config:      cfg,
	}
}

func (m *RateLimiterMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Пропускаем OPTIONS запросы (для CORS preflight)
		if r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Получаем IP адрес клиента
		clientIP := getClientIP(r)

		// Проверяем rate limit
		err := m.rateLimiter.AllowRequest(r.Context(), clientIP)
		if err != nil {
			// При ошибке разрешаем запрос (fail open strategy)
			slog.Warn("Rate limiter error, allowing request", "error", err, "ip", clientIP)
			next.ServeHTTP(w, r)
			return
		}

		if !result.Allowed {
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(m.config.RateLimiter.RequestsPerMinute))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(result.RetryAfter).Unix(), 10))
			w.Header().Set("Retry-After", strconv.FormatFloat(result.RetryAfter.Seconds(), 'f', 0, 64))

			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Устанавливаем headers с информацией о rate limit
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(m.config.RateLimiter.RequestsPerMinute))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))

		next.ServeHTTP(w, r)
	})
}

// Вспомогательная функция для получения реального IP
func getClientIP(r *http.Request) string {
	// Проверяем X-Forwarded-For для случаев за proxy
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Проверяем X-Real-IP
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Берем RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
