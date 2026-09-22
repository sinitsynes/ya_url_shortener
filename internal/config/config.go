package config

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
	Storage       string `env:"STORAGE"`
}

func Load() *Config {
	cfg := &Config{
		ServerAddress: "0.0.0.0:8080",
		BaseURL:       "http://localhost:8080",
		Storage:       "storage.txt",
	}

	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "application address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "shortened base url address")
	flag.StringVar(&cfg.Storage, "s", cfg.Storage, "storage file path")

	flag.Parse()

	_ = env.Parse(cfg)

	return cfg
}
