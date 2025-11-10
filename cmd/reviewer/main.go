package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Caritas-Team/reviewer/internal/config"
	"github.com/Caritas-Team/reviewer/internal/handler"
	"github.com/Caritas-Team/reviewer/internal/logger"
	"github.com/Caritas-Team/reviewer/internal/memecached"
	"github.com/Caritas-Team/reviewer/internal/metrics"
	"github.com/Caritas-Team/reviewer/internal/usecase/file"
)

func createTestData(cache *memecached.Cache) {
	ctx := context.Background()

	// Создаем тестовые файлы
	testFiles := []struct {
		uuid     string
		filename string
		status   string
	}{
		{"test1", "document1.pdf", "NEW"},
		{"test2", "report.pdf", "DOWNLOADED"}, // Этот должен удалиться
		{"test3", "data.pdf", "PROCESSING"},
		{"test4", "final.pdf", "DOWNLOADED"}, // И этот тоже
	}

	for _, tf := range testFiles {
		// Создаем файл
		filePath := "./files/" + tf.uuid + ".pdf"
		os.WriteFile(filePath, []byte("test content"), 0644)

		// Создаем запись в memcached
		metadata := map[string]string{
			"uuid":     tf.uuid,
			"status":   tf.status,
			"filename": tf.filename,
		}
		data, _ := json.Marshal(metadata)
		cache.Set(ctx, tf.uuid, data, time.Hour)

		slog.Info("Created test file", "uuid", tf.uuid, "status", tf.status)
	}
}

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

	fileCleaner := file.NewFileCleaner(cache)

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := fileCleaner.DeleteDownloadedFiles(background); err != nil {
					slog.Error("file cleaner delete error", "err", err)
				} else {
					slog.Info("file cleaner deleted successfully")
				}
			}
		}
	}()

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
