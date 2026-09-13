package config

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS,required"`
	BaseURL       string `env:"BASE_URL,required"`
}

func Load() *Config {
	cfg := Config{}
	err := env.Parse(&cfg)
	if err != nil {
		flag.Parse()
		httpAddr := flag.String("a", "0.0.0.0:8080", "application address")
		redirectBaseURL := flag.String("b", "http://localhost:8080", "shortened base url address")
		cfg.ServerAddress = *httpAddr
		cfg.BaseURL = *redirectBaseURL
	}
	return &cfg
}
