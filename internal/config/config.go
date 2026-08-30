package config

import (
	"flag"
	"ya_url_shortener/internal/config/server"
)

type Config struct {
	HTTPServer server.HTTPServer
	BaseUrl    string
}

var (
	httpAddr        = flag.String("a", "0.0.0.0:8080", "application address")
	redirectBaseURL = flag.String("b", "http://localhost:8080", "shortened base url address")
)

func Load() *Config {
	flag.Parse()
	baseUrl := *redirectBaseURL

	return &Config{
		HTTPServer: server.HTTPServer{URL: *httpAddr},
		BaseUrl:    baseUrl,
	}
}
