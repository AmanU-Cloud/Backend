package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Caritas-Team/reviewer/internal/config"
	"github.com/Caritas-Team/reviewer/internal/handler"
	"github.com/Caritas-Team/reviewer/internal/logger"
	"github.com/Caritas-Team/reviewer/internal/memcached"
	"github.com/Caritas-Team/reviewer/internal/metrics"
	"github.com/Caritas-Team/reviewer/internal/usecase/user"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load error", "err", err)
		return
	}

	ctx := context.Background()
	cache, err := memcached.NewCache(ctx, cfg)
	if err != nil {
		slog.Error("cache initialization failed", "err", err)
		return
	}
	defer func(cache *memcached.Cache) {
		err := cache.Close()
		if err != nil {
			slog.Error("cache close error", "err", err)
		}
	}(cache)

	rateLimiter := user.NewRateLimiter(cache, cfg)

	rateLimitMiddleware := handler.NewRateLimiterMiddleware(rateLimiter)

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("pong"))
	})

	var finalHandler http.Handler = mux
	finalHandler = rateLimitMiddleware.Handler(finalHandler)

	finalHandler = handler.CORS(handler.CORSConfig{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		AllowCredentials: true, // вынести в config.yaml при надобности
		MaxAgeSeconds:    3600, // вынести в config.yaml при надобности
	})(finalHandler)

	metrics.InitMetrics()
	logger.InitGlobalLogger(cfg)

	srv := &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      finalHandler,
		ReadTimeout:  cfg.Server.ReadTimeout(),
		WriteTimeout: cfg.Server.WriteTimeout(),
		IdleTimeout:  5 * time.Minute,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server start failed", "err", err)
		return
	}
}
