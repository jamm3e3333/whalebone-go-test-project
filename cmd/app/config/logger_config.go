package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type LoggerConfig struct {
	Level   string `env:"CONFIG_LOG_LEVEL"`
	DevMode bool   `env:"CONFIG_LOG_DEVEL_MODE"`
}

func CreateLoggerConfig() (LoggerConfig, error) {
	var cfg LoggerConfig
	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}
