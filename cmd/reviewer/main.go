package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Caritas-Team/reviewer/internal/config"
	"github.com/Caritas-Team/reviewer/internal/handler"
	"github.com/Caritas-Team/reviewer/internal/logger"
	"github.com/Caritas-Team/reviewer/internal/memecached"
	"github.com/Caritas-Team/reviewer/internal/metrics"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load error", "err", err)
		return
	}

	background := context.Background()
	cache, err := memecached.NewCache(background, cfg)
	if err != nil {
		slog.Error("cache initialization failed", "err", err)
		return
	}
	defer func(cache *memecached.Cache) {
		err := cache.Close()
		if err != nil {
			slog.Error("cache close error", "err", err)
		}
	}(cache)

	if cache.IsHealthy(background) {
		slog.Info("Memcached is healthy")
	} else {
		slog.Warn("Memcached is unavailable")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("pong"))
	})

	h := handler.CORS(handler.CORSConfig{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		AllowCredentials: true, // вынести в config.yaml при надобности
		MaxAgeSeconds:    3600, // вынести в config.yaml при надобности
	})(mux)

	metrics.InitMetrics()
	logger.InitGlobalLogger(cfg)

	srv := &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      h,
		ReadTimeout:  cfg.Server.ReadTimeout(),
		WriteTimeout: cfg.Server.WriteTimeout(),
		IdleTimeout:  5 * time.Minute,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server start failed", "err", err)
		return
	}
}
