package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/Caritas-Team/reviewer/internal/config"
	"github.com/Caritas-Team/reviewer/internal/handler"
	"github.com/Caritas-Team/reviewer/internal/logger"
	"github.com/Caritas-Team/reviewer/internal/memecached"
	"github.com/Caritas-Team/reviewer/internal/metrics"
)

func main() {
	// Конфиг
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load error", "err", err)
		return
	}

	// Базовый контекст
	ctx := context.Background()

	// Глобальный логер
	logger.InitGlobalLogger(cfg)

	// Контекст, отменяемый по SIGINT/SIGTERM
	rootCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Кэш
	cache, err := memecached.NewCache(rootCtx, cfg)
	if err != nil {
		slog.Error("cache initialization failed", "err", err)
		return
	}

	if cache.IsHealthy(rootCtx) {
		slog.Info("Memcached is healthy")
	} else {
		slog.Warn("Memcached is unavailable")
	}

	// ready + HTTP маршруты для тестов и прочего
	var ready atomic.Bool
	ready.Store(true)

	mux := http.NewServeMux()

	// Для теста CORS. МОЖНО УДАЛЯТЬ
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("pong"))
	})

	// Для проверки, готов ли сервер принимать новый трафик. МОЖНО УДАЛЯТЬ
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if ready.Load() {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
			return
		}
		http.Error(w, "shutting down", http.StatusServiceUnavailable)
	})

	// Для теста Graceful shutdown. МОЖНО УДАЛЯТЬ
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		d := 10 * time.Second
		if s := r.URL.Query().Get("d"); s != "" {
			if parsed, err := time.ParseDuration(s); err == nil {
				d = parsed
			}
		}
		slog.Info("slow begin", "d", d)
		time.Sleep(d)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("slow ok"))
		slog.Info("slow end")
	})

	// Метрики
	metrics.InitMetricsOn(mux)

	// CORS
	h := handler.CORS(handler.CORSConfig{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAgeSeconds:    cfg.CORS.MaxAgeSeconds,
	})(mux)

	// HTTP сервер
	srv := &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      h,
		ReadTimeout:  cfg.Server.ReadTimeout(),
		WriteTimeout: cfg.Server.WriteTimeout(),
		IdleTimeout:  5 * time.Minute,
	}

	// Запуск сервера
	errCh := make(chan error, 1)
	go func() {
		slog.Info("http server listening", "addr", srv.Addr, "pid", os.Getpid())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	// Сигнал или ошибка сервера
	select {
	case <-rootCtx.Done():
		slog.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			slog.Error("server failed", "err", err)
			_ = cache.Close()
			return
		}
	}

	// Graceful shutdown
	ready.Store(false)

	graceTime := 30 * time.Second
	shCtx, cancel := context.WithTimeout(ctx, graceTime)
	defer cancel()

	if err := srv.Shutdown(shCtx); err != nil {
		slog.Error("http shutdown error", "err", err)
	} else {
		slog.Info("http server shutdown complete")
	}

	if err := cache.Close(); err != nil {
		slog.Error("cache close error", "err", err)
	} else {
		slog.Info("cache closed")
	}

	slog.Info("graceful shutdown finished")

}
