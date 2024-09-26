package pgx

import "time"

type Config struct {
	ConnectionURL     string
	LogLevel          string
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	QueryTimeout      time.Duration
	DefaultMaxConns   int32
	DefaultMinConns   int32
	HealthCheckPeriod time.Duration
}
