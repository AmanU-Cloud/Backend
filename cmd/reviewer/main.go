package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Caritas-Team/reviewer/internal/config"
	"github.com/Caritas-Team/reviewer/internal/handler"
	"github.com/Caritas-Team/reviewer/internal/memecached"
	"github.com/Caritas-Team/reviewer/internal/metrics"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load error", "err", err)
		return
	}

	serverAddr := cfg.Server.Addr()
	slog.Info("Starting server", "addr", serverAddr)

	if cfg.Memcached.Enable {
		if len(cfg.Memcached.Servers) == 0 {
			slog.Error("memcached servers is empty")
			return
		}
		ttl := time.Duration(cfg.Memcached.DefaultTTL) * time.Second
		memcachedAddr := cfg.Memcached.Servers
		prefix := cfg.Memcached.KeyPrefix
		memecached.Init(memcachedAddr, ttl, prefix)
		slog.Info("Starting memcached server", "addr", memcachedAddr, "ttl", ttl, "prefix", prefix)

	} else {
		slog.Info("Memcached servers is disabled")
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
