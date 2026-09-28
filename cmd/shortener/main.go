package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"ya_url_shortener/internal/config"
	dbConfig "ya_url_shortener/internal/config/db"
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

// группирующая структура конфигурация хранилки
type repo struct {
	repo    service.Repository
	pool    *pgxpool.Pool
	cleanup func()
}

func newPostgresRepo(ctx context.Context, pg *dbConfig.PostgresConfig) (repo, error) {
	err := db.RunMigrations(pg.DSN)
	if err != nil {
		return repo{}, err
	}
	pool, err := db.InitDB(ctx, pg)
	if err != nil {
		return repo{}, err
	}
	r := postgres.NewStore(pool)
	return repo{
		repo:    r,
		pool:    pool,
		cleanup: pool.Close,
	}, nil
}

func newFileRepo(filePath string) (repo, error) {
	fs, err := filestorage.NewFileStorage(filePath)
	if err != nil {
		return repo{}, err
	}
	r, err := inmemory.NewStore(fs)
	if err != nil {
		_ = fs.Close()
		return repo{}, err
	}
	return repo{
		repo:    r,
		cleanup: func() { _ = fs.Close() },
	}, nil

}

func newInMemoryRepo() (repo, error) {
	r, err := inmemory.NewStore(nil)
	if err != nil {
		return repo{}, err
	}
	return repo{
		repo:    r,
		cleanup: func() {},
	}, nil
}

func configureRepo(ctx context.Context, settings *config.Config) (repo, error) {
	switch {
	case settings.Database.DSN != "":
		return newPostgresRepo(ctx, settings.Database)
	case settings.FileStoragePath != "":
		return newFileRepo(settings.FileStoragePath)
	default:
		return newInMemoryRepo()
	}
}

func run(logger *slog.Logger) error {
	settings, err := config.Load()
	if err != nil {
		return err
	}
	ctx := context.Background()
	repo, err := configureRepo(ctx, settings)
	if err != nil {
		return err
	}
	defer repo.cleanup()

	// сокращение и хранение URL
	controller := service.NewResourceController(repo.repo)
	rHandler := resource.NewResourceHandler(settings.BaseURL, controller, logger)
	// хэлсчек БД
	hHandler := healthcheck.NewHandler(repo.pool)
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
