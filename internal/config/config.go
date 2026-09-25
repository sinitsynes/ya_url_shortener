package config

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func Load() (*Config, error) {
	cfg := &Config{
		ServerAddress:   "0.0.0.0:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "storage.txt",
	}

	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "application address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "shortened base url address")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "storage file path")

	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
