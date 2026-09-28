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
	filestorage "ya_url_shortener/internal/repository/file_storage"
	inmemory "ya_url_shortener/internal/repository/in_memory"
	"ya_url_shortener/internal/repository/postgres"
	"ya_url_shortener/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// configureRepo собирает хранилку исходя из конфигурации сервиса.
// Если передано подключение к БД, то она выступает основным хранилищем.
// В противном случае собирается inMemory хранение в мапе + файловое хранилище для персистентности.
// Если файл не передан, то остается только inMemory мапа.
// Функция возвращает репозиторий, соединение с БД (для хелсчека), функцию закрытия репозитория, ошибку.
func configureRepo(ctx context.Context, settings *config.Config) (service.Repository, *pgxpool.Pool, func(), error) {
	noop := func() {}

	// БД
	if settings.Database.DSN != "" {
		err := db.RunMigrations(settings.Database.DSN)
		if err != nil {
			return nil, nil, noop, err
		}
		pool, err := db.InitDB(ctx, settings.Database)
		if err != nil {
			return nil, nil, noop, err
		}
		repo := postgres.NewStore(pool)
		cleanup := func() { pool.Close() }
		return repo, pool, cleanup, nil
	}
	// inmemory + file storage
	if settings.FileStoragePath != "" {
		fStorage, err := filestorage.NewFileStorage(settings.FileStoragePath)
		if err != nil {
			return nil, nil, noop, err
		}
		repo, err := inmemory.NewStore(fStorage)
		if err != nil {
			if storageErr := fStorage.Close(); storageErr != nil {
				return nil, nil, noop, storageErr
			}
			return nil, nil, noop, err
		}
		cleanup := func() { _ = fStorage.Close() }
		return repo, nil, cleanup, nil
	}
	// только inmemory
	repo, err := inmemory.NewStore(nil)
	if err != nil {
		return nil, nil, noop, err
	}
	return repo, nil, noop, nil
}

func run(logger *slog.Logger) error {
	settings, err := config.Load()
	if err != nil {
		return err
	}
	ctx := context.Background()
	repo, pool, cleanup, err := configureRepo(ctx, settings)
	if err != nil {
		return err
	}
	defer cleanup()

	// сокращение и хранение URL
	controller := service.NewResourceController(repo)
	rHandler := resource.NewResourceHandler(settings.BaseURL, controller, logger)
	// хэлсчек БД
	hHandler := healthcheck.NewHandler(pool)
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
