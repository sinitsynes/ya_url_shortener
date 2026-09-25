package repository

import (
	"context"

	"ya_url_shortener/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func (pg *PostgresStore) CreateResource(_ model.Resource) (model.Resource, error) {
	return model.Resource{}, ErrNotImplemented
}

func (pg *PostgresStore) GetResourceByID(_ int32) (model.Resource, error) {
	return model.Resource{}, ErrNotImplemented
}

func (pg *PostgresStore) GetResourceByURL(_ string) (model.Resource, error) {
	return model.Resource{}, ErrNotImplemented
}

func (pg *PostgresStore) Ping(ctx context.Context) error {
	return pg.pool.Ping(ctx)
}
