package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"ya_url_shortener/internal/config"
	"ya_url_shortener/internal/handler"
	"ya_url_shortener/internal/infra/db"
	"ya_url_shortener/internal/infra/httpserver"
	"ya_url_shortener/internal/repository"
	"ya_url_shortener/internal/service"
)

func run(logger *slog.Logger) error {
	settings, err := config.Load()
	if err != nil {
		return err
	}
	repo, err := repository.NewStore(settings.FileStoragePath)
	if err != nil {
		return err
	}
	defer repo.Close()
	ctx := context.Background()
	dbPool, err := db.InitDB(ctx, settings.Database.DSN)
	if err != nil {
		return err
	}

	controller := service.NewResourceController(repo)
	h := handler.NewResourceHandler(settings.BaseURL, controller, logger, dbPool)
	router := handler.NewRouter(h, logger)
	server := httpserver.NewServer(settings.ServerAddress, router)
	logger.Info("server started", "address", settings.ServerAddress)
	return server.ListenAndServe()
}

func main() {
	logger := config.NewLogger()
	slog.SetDefault(logger)
	if err := run(logger); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("fatal error", "error", err)
		os.Exit(1)
	}
}
