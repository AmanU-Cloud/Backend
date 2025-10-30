package logger

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"

	yaml "gopkg.in/yaml.v3"
)

var GlobalLogger *slog.Logger

// Структура для представления конфигурации
type Config struct {
	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"logging"`
}

// Чтение и обработка конфигурации
func loadConfig(path string) (*Config, error) {
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

	var cfg Config
	err = yaml.Unmarshal(buf.Bytes(), &cfg)
	if err != nil {
		return nil, fmt.Errorf("ошибка при разборе конфигурации: %w", err)
	}

	return &cfg, nil
}

func init() {
	cfg, err := loadConfig("../cfg/config.yml") // Относительный путь к файлу конфигурации
	if err != nil {
		slog.Error("Ошибка при загрузке (использованы настройки по умолчанию):", slog.Any("error", err))
		// Используем стандартные настройки
		cfg = &Config{
			Logging: struct {
				Level  string `yaml:"level"`
				Format string `yaml:"format"`
			}{
				Level:  "info",
				Format: "json", // Значение по умолчанию
			},
		}
	}

	// Определяем уровень логирования
	var level slog.Level
	switch cfg.Logging.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		slog.Warn("Указанный уровень логирования не распознается, используется Debug по умолчанию.")
		level = slog.LevelDebug // Устанавливаем уровень по умолчанию
	}

	// Определяем формат логирования
	var handler slog.Handler
	switch cfg.Logging.Format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     level,
		})
	case "text": // Возможно, поддержка текстового формата
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     level,
		})
	default:
		slog.Warn("Формат логирования не найден, использован JSON по умолчанию")
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     level,
		})
	}

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
