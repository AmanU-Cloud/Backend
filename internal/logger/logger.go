package logger

import (
	"log/slog"
	"os"

	"sync"

	"github.com/Caritas-Team/reviewer/internal/config"
)

type Logger struct {
	mu     sync.Mutex
	logger *slog.Logger
}

// Создание логгера
func NewLogger(cfg config.Config) *Logger {
	// Локальные переменные для уровня и формата логирования
	var localLevel string
	var localFormat string

	// Проверка наличия уровня и формата
	if cfg.Logging.Level != "" {
		localLevel = cfg.Logging.Level
	} else {
		slog.Warn("Отсутствует уровень логирования, используетя debug")
		localLevel = "debug"
	}

	if cfg.Logging.Format != "" {
		localFormat = cfg.Logging.Format
	} else {
		slog.Warn("Отсутствует формат логирования, используется json")
		localFormat = "json"
	}

	// Проверяем корректность уровня логирования
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[localLevel] {
		slog.Warn("Некорректный уровень логирования в конфигурации, используем 'debug'.", "provided_level", localLevel)
		localLevel = "debug"
	}

	// Проверка корректности формата логирования
	validFormats := map[string]bool{"json": true, "text": true}
	if !validFormats[localFormat] {
		slog.Warn("Некорректный формат логирования в конфигурации, используем 'json'.", "provided_format", localFormat)
		localFormat = "json"
	}

	// Определяем уровень логирования
	var level slog.Level
	switch localLevel {
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelDebug
	}

	// Определяем формат логирования
	var handler slog.Handler
	if localFormat == "text" {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     level,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     level,
		})
	}

	return &Logger{
		logger: slog.New(handler),
	}
}

// Глобальная переменная логгера
var GlobalLogger *Logger

// Инициализация глобального логгера
func InitGlobalLogger(cfg config.Config) {
	GlobalLogger = NewLogger(cfg)
}

// Методы логирования с мьютексом
func (l *Logger) Info(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Info(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Warn(msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Debug(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Error(msg, args...)
}
