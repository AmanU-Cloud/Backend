package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type HTTP struct {
	Address string `mapstructure:"address"` // ":8080"
}

type CORS struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAgeSeconds    int      `mapstructure:"max_age_seconds"`
}

type Config struct {
	HTTP HTTP `mapstructure:"http"`
	CORS CORS `mapstructure:"cors"`
}

// Defaults —  дефолты, если чего-то нет в конфиге.
func Defaults() Config {
	return Config{
		HTTP: HTTP{Address: ":8080"},
		CORS: CORS{
			AllowedOrigins:   nil, // пусто = AllowAll
			AllowedMethods:   nil, // дефолты из handler.DefaultCORSMethods
			AllowedHeaders:   nil, // дефолты из handler.DefaultCORSHeaders
			AllowCredentials: true,
			MaxAgeSeconds:    int((time.Hour).Seconds()),
		},
	}
}

// Load загружает конфиг из config/config.yaml.
func Load() (Config, error) {
	var cfg Config

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("config") // переход в папку config. Если конфиг будет переноситься, надо исправить

	if err := v.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal config: %w", err)
	}
	return cfg, nil
}
