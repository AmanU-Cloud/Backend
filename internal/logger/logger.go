package logger

import (
	"log/slog"
	"os"
)

var GlobalLogger *slog.Logger

func init() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	})
	GlobalLogger = slog.New(handler)
	slog.SetDefault(GlobalLogger)
}

// Методы логирования

// Info записывает инфо-сообщение
func Info(msg string, args ...any) {
	GlobalLogger.Info(msg, args...)
}

// Warn записывает предупреждение
func Warn(msg string, args ...any) {
	GlobalLogger.Warn(msg, args...)
}

// Debug записывает дебаг-сообщение
func Debug(msg string, args ...any) {
	GlobalLogger.Debug(msg, args...)
}

// Error записывает ошибку
func Error(msg string, args ...any) {
	GlobalLogger.Error(msg, args...)
}
