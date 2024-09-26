package config

import (
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type APPConfig struct {
	Port               int32         `env:"CONFIG_HTTP_LISTEN_PORT" env-default:"8080"`
	AppEnv             string        `env:"APP_ENV" env-default:"development"`
	AllowOrigins       string        `env:"CONFIG_ALLOW_ORIGINS"`
	ReadTimeout        time.Duration `env:"CONFIG_HTTP_READ_TIMEOUT"`
	WriteTimeout       time.Duration `env:"CONFIG_HTTP_WRITE_TIMEOUT"`
	ShutdownTimeout    time.Duration `env:"CONFIG_HTTP_SHUTDOWN_TIMEOUT"`
	HealthCheckTimeout time.Duration `env:"CONFIG_HEALTH_CHECK_TIMEOUT"`
	Timezone           string        `env:"CONFIG_TIMEZONE"`
	AppName            string        `env:"CONFIG_APP_NAME" env-default:"whalebone_clients"`
}

func CreateAPPConfig() (APPConfig, error) {
	var cfg APPConfig
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (cfg APPConfig) AllowedOrigins() []string {
	return strings.Split(cfg.AllowOrigins, ";")
}
