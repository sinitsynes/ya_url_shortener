package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	"ya_url_shortener/internal/config"
	"ya_url_shortener/internal/handler"
	"ya_url_shortener/internal/infra/httpserver"
	"ya_url_shortener/internal/repository"
	"ya_url_shortener/internal/service"
)

func run() error {
	settings, err := config.Load()
	if err != nil {
		return err
	}
	logger := config.NewLogger()
	slog.SetDefault(logger)

	repo, err := repository.NewStore(settings.FileStoragePath)
	defer repo.Close()
	if err != nil {
		return err
	}
	controller := service.NewResourceController(repo)
	h := handler.NewResourceHandler(settings.BaseURL, controller, logger)
	router := handler.NewRouter(h, logger)
	server := httpserver.NewServer(settings.ServerAddress, router)
	logger.Info("server started", "address", settings.ServerAddress)
	return server.ListenAndServe()
}

func main() {
	if err := run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("fatal error", "error", err)
		os.Exit(1)
	}
}
