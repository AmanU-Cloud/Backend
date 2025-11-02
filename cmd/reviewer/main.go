package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Caritas-Team/reviewer/internal/config"
	"github.com/Caritas-Team/reviewer/internal/handler"
	"github.com/Caritas-Team/reviewer/internal/metrics"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load error", "err", err)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("pong"))
	})

	h := handler.CORS(handler.CORSConfig{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods, // nil -> дефолты
		AllowedHeaders:   cfg.CORS.AllowedHeaders, // nil -> дефолты
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAgeSeconds:    cfg.CORS.MaxAgeSeconds,
	})(mux)

	metrics.InitMetrics()

	srv := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      h,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  5 * time.Minute,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server start failed", "err", err)
		return
	}
}
