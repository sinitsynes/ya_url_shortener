package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"ya_url_shortener/internal/config"
	"ya_url_shortener/internal/handler"
	"ya_url_shortener/internal/handler/healthcheck"
	"ya_url_shortener/internal/handler/resource"
	"ya_url_shortener/internal/infra/db"
	"ya_url_shortener/internal/infra/httpserver"
	in_memory "ya_url_shortener/internal/repository/in_memory"
	"ya_url_shortener/internal/repository/postgres"
	"ya_url_shortener/internal/service"

	"github.com/go-chi/chi/v5"
)

// configureRepo возвращает репозиторий и пингер для работы с БД.
// Пингер это или пул БД, который можно пингануть, или noop для in-memory хранилки.
func configureRepo(ctx context.Context, settings *config.Config) (service.Repository, db.Pinger, error) {
	switch {
	case settings.Database.DSN != "":
		return postgres.NewPostgresStore(ctx, settings.Database)
	case settings.FileStoragePath != "":
		repo, err := in_memory.NewWithFileStorage(settings.FileStoragePath)
		return repo, db.UnavailablePinger{}, err
	default:
		repo, err := in_memory.NewInMemoryStore()
		return repo, db.UnavailablePinger{}, err
	}
}

func run(logger *slog.Logger) error {
	settings, err := config.Load()
	if err != nil {
		return err
	}
	ctx := context.Background()
	repository, pinger, err := configureRepo(ctx, settings)
	if err != nil {
		return err
	}
	defer repository.Close()

	// сокращение и хранение URL
	controller := service.NewResourceController(settings.BaseURL, repository)
	rHandler := resource.NewResourceHandler(controller, logger)
	// хэлсчек БД
	hHandler := healthcheck.NewHandler(pinger, logger)
	baseRouter := handler.NewRouter(logger,
		func(r chi.Router) { healthcheck.RegisterRoutes(r, hHandler) },
		func(r chi.Router) { resource.RegisterRoutes(r, rHandler) },
	)
	server := httpserver.NewServer(settings.ServerAddress, baseRouter)
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
