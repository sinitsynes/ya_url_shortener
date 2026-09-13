package config

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

func Load() *Config {
	cfg := &Config{
		ServerAddress: "0.0.0.0:8080",
		BaseURL:       "http://localhost:8080",
	}

	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "application address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "shortened base url address")
	flag.Parse()

	_ = env.Parse(cfg)

	return cfg
}
