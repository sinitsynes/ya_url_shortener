package postgres

import (
	"context"

	dbConfig "ya_url_shortener/internal/config/db"
	"ya_url_shortener/internal/infra/db"
)

func NewPostgresStore(ctx context.Context, pg *dbConfig.PostgresConfig) (*Store, db.Pinger, error) {
	err := db.RunMigrations(pg.DSN)
	if err != nil {
		return nil, nil, err
	}
	pool, err := db.InitDB(ctx, pg)
	if err != nil {
		return nil, nil, err
	}
	return NewStore(pool), pool, nil
}
