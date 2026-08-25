package config

import (
	"flag"
	"strings"
)

type HTTPServer struct {
	Host string
	Port string
}

type Config struct {
	HTTPServer HTTPServer
	BaseUrl    string
}

var (
	addr    = flag.String("a", "0.0.0.0:8080", "application address")
	baseURL = flag.String("b", "http://localhost:8080", "shortened base url address")
)

func Load() *Config {
	flag.Parse()
	parts := strings.Split(*addr, ":")
	return &Config{
		HTTPServer: HTTPServer{Host: parts[0], Port: parts[1]},
		BaseUrl:    *baseURL,
	}
}
