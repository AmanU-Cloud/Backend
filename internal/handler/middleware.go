package handler

import (
	"net/http"

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
