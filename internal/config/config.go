package config

import (
	"flag"

	"ya_url_shortener/internal/config/db"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerAddress   string        `env:"SERVER_ADDRESS"`
	BaseURL         string        `env:"BASE_URL"`
	FileStoragePath string        `env:"FILE_STORAGE_PATH"`
	SecretKey       AuthSecretKey `env:"SECRET_KEY"`
	Database        *db.PostgresConfig
}

func Load() (*Config, error) {
	var secretKey string
	cfg := &Config{
		ServerAddress:   "0.0.0.0:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
		Database:        &db.PostgresConfig{DSN: ""},
		SecretKey:       AuthSecretKey(secretKey),
	}

	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "application address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "shortened base url address")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "storage file path")
	flag.StringVar(&cfg.Database.DSN, "d", cfg.Database.DSN, "database DSN")
	flag.StringVar(&secretKey, "s", secretKey, "secret key")

	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
