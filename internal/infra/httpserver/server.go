package httpserver

import (
	"net/http"
)

func NewServer(address string, mux http.Handler) *http.Server {
	s := http.Server{Addr: address, Handler: mux}
	return &s
}
