package postgres

import (
	"context"

	"ya_url_shortener/internal/model"

	storage "ya_url_shortener/internal/repository/postgres/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool    *pgxpool.Pool
	queries *storage.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		pool:    pool,
		queries: storage.New(pool),
	}
}

func (pg *Store) CreateResource(ctx context.Context, r model.Resource) (model.Resource, error) {
	createParams := storage.CreateResourceParams{OriginalUrl: r.Address, ShortenedUrl: r.Shortened}
	created, err := pg.queries.CreateResource(ctx, createParams)
	if err != nil {
		return model.Resource{}, err
	}
	return model.Resource{
		Address:   created.OriginalUrl,
		Shortened: created.ShortenedUrl,
		ID:        created.ID,
	}, nil
}

func (pg *Store) GetResourceByID(ctx context.Context, id int32) (model.Resource, error) {
	item, err := pg.queries.GetResourceByID(ctx, id)
	if err != nil {
		return model.Resource{}, err
	}
	return model.Resource{
		ID:        item.ID,
		Address:   item.OriginalUrl,
		Shortened: item.ShortenedUrl,
	}, nil
}

func (pg *Store) GetResourceByURL(ctx context.Context, url string) (model.Resource, error) {
	item, err := pg.queries.GetResourceByURL(ctx, url)
	if err != nil {
		return model.Resource{}, err
	}
	return model.Resource{
		ID:        item.ID,
		Address:   item.OriginalUrl,
		Shortened: item.ShortenedUrl,
	}, nil
}

func (pg *Store) Ping(ctx context.Context) error {
	return pg.pool.Ping(ctx)
}
