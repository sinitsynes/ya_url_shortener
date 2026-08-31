package main

import (
	"errors"
	"log"
	"net/http"
	"ya_url_shortener/internal/config"
	"ya_url_shortener/internal/handler"
	"ya_url_shortener/internal/infra/httpserver"
	"ya_url_shortener/internal/repository"
	"ya_url_shortener/internal/service"
)

func run() error {
	config := config.Load()
	repo := repository.NewStore()
	controller := service.NewResourceController(repo)
	h := handler.NewResourceHandler(config.BaseURL, controller)
	router := handler.NewRouter(h)
	server := httpserver.NewServer(config.HTTPServer.URL, router)
	err := server.ListenAndServe()
	return err
}

func main() {
	err := run()
	if errors.Is(err, http.ErrServerClosed) {
		log.Fatal("server closed unexpectedly: %w", err)
	}
}
