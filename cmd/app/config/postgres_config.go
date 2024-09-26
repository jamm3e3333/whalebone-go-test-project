package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type PostgresConfig struct {
	Host              string        `env:"CONFIG_DATABASE_HOST"`
	Port              int32         `env:"CONFIG_DATABASE_PORT"`
	User              string        `env:"CONFIG_DATABASE_USER"`
	Password          string        `env:"CONFIG_DATABASE_PASSWORD"`
	LogLevel          string        `env:"CONFIG_LOG_LEVEL"`
	DBName            string        `env:"CONFIG_DATABASE_NAME"`
	MaxConnLifetime   time.Duration `env:"CONFIG_DATABASE_POOL_MAX_CONN_LIFETIME"`
	MaxConnIdleTIme   time.Duration `env:"CONFIG_DATABASE_POOL_MAX_CONN_IDLE_TIME"`
	QueryTimeout      time.Duration `env:"CONFIG_DATABASE_QUERY_TIMEOUT"`
	MaxConns          int32         `env:"CONFIG_DATABASE_POOL_MAX_CONNS"`
	MinConns          int32         `env:"CONFIG_DATABASE_POOL_MIN_CONNS"`
	HealthCheckPeriod time.Duration `env:"CONFIG_DATABASE_POOL_HEALTH_CHECK_PERIOD"`
	SSLMode           string        `env:"CONFIG_DATABASE_SSL_MODE" env-default:"disable"`
}

func CreatePostgresConfig() (PostgresConfig, error) {
	var cfg PostgresConfig
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}

func (pc PostgresConfig) ConnectionURL() string {
	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		pc.User,
		pc.Password,
		pc.Host,
		pc.Port,
		pc.DBName,
		pc.SSLMode,
	)
}
