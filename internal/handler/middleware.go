package handler

import (
	"net"
	"net/http"
	"strings"

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
	limiter *user.RateLimiter
}

func NewRateLimiterMiddleware(limiter *user.RateLimiter) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		limiter: limiter,
	}
}

func (m *RateLimiterMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		if host, _, err := net.SplitHostPort(ip); err == nil {
			ip = host
		}

		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ips := strings.Split(forwarded, ",")
			if len(ips) > 0 {
				ip = strings.TrimSpace(ips[0])
			}
		}

		err := m.limiter.AllowRequest(r.Context(), ip)
		if err != nil {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
