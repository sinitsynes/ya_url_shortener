package main

import (
	"net/http"
	"ya_url_shortener/internal/config"
	"ya_url_shortener/internal/handler"
	"ya_url_shortener/internal/infra/httpserver"
)

func main() {
	c := config.Load()
	h := handler.NewHandler()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", h.CreateUrl)
	mux.HandleFunc("GET /{id}", h.GetUrl)
	server := httpserver.NewServer(c, mux)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}

}
