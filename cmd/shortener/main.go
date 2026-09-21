package main

import (
	"errors"
	"log"
	"log/slog"
	"net/http"

	"ya_url_shortener/internal/config"
	"ya_url_shortener/internal/handler"
	"ya_url_shortener/internal/infra/httpserver"
	"ya_url_shortener/internal/repository"
	"ya_url_shortener/internal/service"
)

func run() error {
	settings := config.Load()
	logger := config.NewLogger()
	slog.SetDefault(logger)

	repo := repository.NewStore()
	controller := service.NewResourceController(repo)
	h := handler.NewResourceHandler(settings.BaseURL, controller, logger)
	router := handler.NewRouter(h)
	server := httpserver.NewServer(settings.ServerAddress, router)
	err := server.ListenAndServe()
	return err
}

func main() {
	err := run()
	if errors.Is(err, http.ErrServerClosed) {
		log.Fatal("server closed unexpectedly: %w", err)
	}
}
