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
	slog.Info("server started", "address", settings.ServerAddress)
	return server.ListenAndServe()
}

func main() {
	if err := run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
