package logger

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/Caritas-Team/reviewer/cfg"
	yaml "gopkg.in/yaml.v3"
)

type Logger struct {
	mu     sync.Mutex
	logger *slog.Logger
}

// Чтение и обработка конфигурации
func loadConfig(path string) (*cfg.Config, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("ошибка при открытии файла конфигурации: %w", err)
	}
	defer fd.Close()

	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(fd)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении файла конфигурации: %w", err)
	}

	var cfg cfg.Config
	err = yaml.Unmarshal(buf.Bytes(), &cfg)
	if err != nil {
		return nil, fmt.Errorf("ошибка при разборе конфигурации: %w", err)
	}

	return &cfg, nil
}

// Создание логгера
func NewLogger(configPath string) *Logger {
	config, err := loadConfig(configPath)
	if err != nil {
		slog.Error("Ошибка при загрузке (использованы настройки по умолчанию):", slog.Any("error", err))
		// Используем стандартные настройки
		config = &cfg.Config{
			Logging: struct {
				Level  string `yaml:"level"`
				Format string `yaml:"format"`
			}{
				Level:  "debug",
				Format: "json", // Значение по умолчанию
			},
		}
	}

	// Проверяем корректность уровня логирования
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLevels[config.Logging.Level] {
		slog.Warn("Некорректный уровень логирования в конфигурации, установлен 'debug'", slog.Any("provided_level", config.Logging.Level))
		config.Logging.Level = "debug"
	}

	// Проверка корректности формата логирования
	validFormats := map[string]bool{
		"json": true,
		"text": true,
	}

	if !validFormats[config.Logging.Format] {
		slog.Warn("Некорректный формат логирования в конфигурации, установлен 'json'", slog.Any("provided_format", config.Logging.Format))
		config.Logging.Format = "json"
	}

	// Определяем уровень логирования
	var level slog.Level
	switch config.Logging.Level {
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
	if config.Logging.Format == "text" {
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

	logger := slog.New(handler)

	return &Logger{
		logger: logger,
	}
}

// Глобальная переменная логгера
var GlobalLogger *Logger

// Инициализация глобального логгера
func InitGlobalLogger(configPath string) {
	GlobalLogger = NewLogger(configPath)
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
