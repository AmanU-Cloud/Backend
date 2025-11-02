package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Caritas-Team/reviewer/internal/metrics"
  "github.com/Caritas-Team/reviewer/internal/logger"
	"github.com/google/uuid" //добавил и использовал, чтобы появился go.sum
)

func main() {
	_ = uuid.New()
  logger.InitGlobalLogger("../../cfg/config.yml")
	metrics.InitMetrics()

	srv := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  time.Second * 10, // Таймаут на чтение запроса
		WriteTimeout: time.Second * 10, // Таймаут на запись ответа
		IdleTimeout:  time.Minute * 5,  // Максимальное время ожидания неактивного подключения
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Ошибка запуска сервера", "err", err)
		return
	}
}
