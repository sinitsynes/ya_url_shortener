package db

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" //nolint:blank-imports
	_ "github.com/golang-migrate/migrate/v4/source/file"     //nolint:blank-imports
	"github.com/jackc/pgx/v5/pgxpool"
)

const ConnectionTimeout time.Duration = 1 * time.Second

func InitDB(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(ctx, ConnectionTimeout)
	defer cancel()
	return pgxpool.New(ctx, connString)
}

func RunMigrations(connString string) error {
	u, parseErr := url.Parse(connString)
	if parseErr != nil {
		return parseErr
	}
	u.Scheme = "pgx5"
	dsn := u.String()
	migration, migrationErr := migrate.New("file://./migrations", dsn)
	if migrationErr != nil {
		return migrationErr
	}
	defer migration.Close()

	if applyMigrationErr := migration.Up(); applyMigrationErr != nil && !errors.Is(applyMigrationErr, migrate.ErrNoChange) {
		return applyMigrationErr
	}
	return nil
}
