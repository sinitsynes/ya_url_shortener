package db

type PostgresConfig struct {
	DSN string `env:"DATABASE_DSN"`
}
