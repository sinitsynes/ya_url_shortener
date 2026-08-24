package httpserver

import (
	"net"
	"net/http"
	"ya_url_shortener/internal/config"
)

func NewServer(config *config.Config, mux http.Handler) *http.Server {
	addr := net.JoinHostPort(config.HTTPServer.Host, config.HTTPServer.Port)
	s := http.Server{Addr: addr, Handler: mux}
	return &s
}
