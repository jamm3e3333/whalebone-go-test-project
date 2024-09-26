package helper

import (
	"time"
)

type PostgresConfig struct{}

func NewPostgresConfig() *PostgresConfig {
	return &PostgresConfig{}
}

func (tpg *PostgresConfig) ConnectionURL() string {
	connURL := "postgres://postgres:postgres@postgres/whalebone-clients?sslmode=disable"
	if AllowDebug() {
		connURL = "postgres://postgres:postgres@localhost:54320/whalebone-clients?sslmode=disable"
	}

	return connURL
}

func (tpg *PostgresConfig) LogLevel() string {
	return "debug"
}

func (tpg *PostgresConfig) MaxConnLifetime() time.Duration {
	return 10 * time.Second
}

func (tpg *PostgresConfig) MaxConnIdleTime() time.Duration {
	return 10 * time.Second
}

func (tpg *PostgresConfig) DefaultMaxConns() int32 {
	return 10
}

func (tpg *PostgresConfig) DefaultMinConns() int32 {
	return 1
}

func (tpg *PostgresConfig) HealthCheckPeriod() time.Duration {
	return 10 * time.Second
}

func (tpg *PostgresConfig) QueryTimeout() time.Duration {
	return 5 * time.Second
}

func (tpg *PostgresConfig) AppName() string {
	return "testing_app"
}
