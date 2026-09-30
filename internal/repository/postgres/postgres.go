package postgres

import (
	"context"
	"errors"

	"ya_url_shortener/internal/model"
	"ya_url_shortener/internal/repository"

	storage "ya_url_shortener/internal/repository/postgres/sqlc"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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
	createParams := storage.CreateResourceParams{
		OriginalUrl: r.OriginalURL,
		ShortUrl:    r.ShortURL}
	created, err := pg.queries.CreateResource(ctx, createParams)
	if err != nil {
		return model.Resource{}, err
	}
	return model.Resource{
		OriginalURL: created.OriginalUrl,
		ShortURL:    created.ShortUrl,
		ID:          created.ID,
	}, nil
}

func (pg *Store) GetResourceByID(ctx context.Context, id int32) (model.Resource, error) {
	item, err := pg.queries.GetResourceByID(ctx, id)
	if err != nil {
		return model.Resource{}, err
	}
	return model.Resource{
		ID:          item.ID,
		OriginalURL: item.OriginalUrl,
		ShortURL:    item.ShortUrl,
	}, nil
}

func (pg *Store) GetResourceByShortURL(ctx context.Context, url string) (model.Resource, error) {
	item, err := pg.queries.GetResourceByShortURL(ctx, url)
	if err != nil {
		return model.Resource{}, err
	}
	return model.Resource{
		ID:            item.ID,
		OriginalURL:   item.OriginalUrl,
		ShortURL:      item.ShortUrl,
		CorrelationID: item.CorrelationID,
	}, nil
}

func (pg *Store) GetResourceByOriginalURL(ctx context.Context, url string) (model.Resource, error) {
	item, err := pg.queries.GetResourceByOriginalURL(ctx, url)
	if err != nil {
		return model.Resource{}, err
	}
	return model.Resource{
		ID:            item.ID,
		OriginalURL:   item.OriginalUrl,
		ShortURL:      item.ShortUrl,
		CorrelationID: item.CorrelationID,
	}, nil
}

func (pg *Store) CreateBatch(ctx context.Context, resources []model.Resource) ([]model.Resource, error) {
	params := storage.CreateBatchParams{}
	for _, r := range resources {
		params.OriginalUrls = append(params.OriginalUrls, r.OriginalURL)
		params.ShortUrls = append(params.ShortUrls, r.ShortURL)
		params.CorrelationIds = append(params.CorrelationIds, r.CorrelationID)
	}
	items, err := pg.queries.CreateBatch(ctx, params)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return nil, repository.ErrConflict
			}
		}
		return nil, err
	}
	out := make([]model.Resource, 0, len(items))
	for i, item := range out {
		out[i] = model.Resource{
			ID:            item.ID,
			OriginalURL:   item.OriginalURL,
			ShortURL:      item.ShortURL,
			CorrelationID: item.CorrelationID,
		}
	}
	return out, nil
}
