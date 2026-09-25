package httpserver

import (
	"net/http"
	"time"
)

const (
	readHeaderTimeoutSeconds = 3
	readTimeoutSeconds       = 15
)

func NewServer(address string, mux http.Handler) *http.Server {
	s := http.Server{
		Addr:              address,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeoutSeconds * time.Second,
		ReadTimeout:       readTimeoutSeconds * time.Second}
	return &s
}
