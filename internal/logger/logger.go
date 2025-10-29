package logger

import (
	"log/slog"
	"os"
)

func init() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,           // Место возникновения ошибки
		Level:     slog.LevelInfo, // Уровень логирования
	})
	globalLogger := slog.New(handler)
	slog.SetDefault(globalLogger)
}
