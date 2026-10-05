package db

import (
	"context"
	"errors"
	"net/url"

	"ya_url_shortener/internal/config/db"
	"ya_url_shortener/migrations"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" //nolint:blank-imports
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDB(ctx context.Context, dbConfig *db.PostgresConfig) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dbConfig.DSN)
	if err != nil {
		return nil, err
	}
	config.ConnConfig.ConnectTimeout = dbConfig.ConnectionTimeout

	ctx, cancel := context.WithTimeout(ctx, dbConfig.ConnectionTimeout)
	defer cancel()
	return pgxpool.NewWithConfig(ctx, config)
}

func RunMigrations(connString string) error {
	migrationsFolder, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}

	u, parseErr := url.Parse(connString)
	if parseErr != nil {
		return parseErr
	}
	u.Scheme = "pgx5"
	dsn := u.String()
	migration, migrationErr := migrate.NewWithSourceInstance("iofs", migrationsFolder, dsn)
	if migrationErr != nil {
		return migrationErr
	}
	defer migration.Close()

	if applyMigrationErr := migration.Up(); applyMigrationErr != nil &&
		!errors.Is(applyMigrationErr, migrate.ErrNoChange) {
		return applyMigrationErr
	}
	return nil
}
