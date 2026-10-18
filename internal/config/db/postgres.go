package db

import (
	"time"
)

const ConnectionTimeout time.Duration = 5 * time.Second

type PostgresConfig struct {
	DSN               string `env:"DATABASE_DSN"`
	ConnectionTimeout time.Duration
}
