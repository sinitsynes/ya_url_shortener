package service

import (
	"context"
	"errors"

	"ya_url_shortener/internal/model"
	"ya_url_shortener/internal/repository"
	"ya_url_shortener/pkg/encoder"
)

const maxCreateAttempts int32 = 5

var ErrMaxRetriesExceeded = errors.New("failed to generate unique shortened url")

type Repository interface {
	CreateResource(context.Context, model.Resource) (model.Resource, error)
	GetResourceByID(context.Context, int32) (model.Resource, error)
	GetResourceByURL(context.Context, string) (model.Resource, error)
	CreateBatch(context.Context, []model.Resource) ([]model.ResourceBatchOutput, error)
}

type Controller struct {
	baseURL string
	store   Repository
}

func NewResourceController(baseURL string, repository Repository) *Controller {
	return &Controller{
		baseURL: baseURL,
		store:   repository}
}

// withConflictRetry повторяет попытку создания записей, пока количество попыток не перевалит за константу.
// При условии, что срабатывать будет только ошибка ErrConflict.
func withConflictRetry[T any](attempt func(salt int32) (T, error)) (T, error) {
	var zero T
	for salt := range maxCreateAttempts {
		result, err := attempt(salt)
		if err == nil {
			return result, nil
		}
		if !errors.Is(err, repository.ErrConflict) {
			return zero, err
		}
	}
	return zero, ErrMaxRetriesExceeded
}

func (s *Controller) CreateResource(ctx context.Context, originalURL string) (model.Resource, error) {
	return withConflictRetry(func(salt int32) (model.Resource, error) {
		created, err := s.store.CreateResource(ctx, model.Resource{
			OriginalURL: originalURL,
			ShortURL:    encoder.EncodeURL(originalURL, salt),
		})
		if err != nil {
			return model.Resource{}, err
		}
		created.ShortURL = s.baseURL + "/" + created.ShortURL
		return created, nil
	})
}

func (s *Controller) GetResource(ctx context.Context, shortenedURL string) (model.Resource, error) {
	r, err := s.store.GetResourceByURL(ctx, shortenedURL)
	if err != nil {
		return model.Resource{}, err
	}
	r.ShortURL = s.baseURL + "/" + r.ShortURL
	return r, nil
}

func (s *Controller) CreateBatch(
	ctx context.Context,
	batch []model.ResourceBatchInput,
) ([]model.ResourceBatchOutput, error) {
	return withConflictRetry(func(salt int32) ([]model.ResourceBatchOutput, error) {
		resources := make([]model.Resource, len(batch))
		for i, item := range batch {
			resources[i] = model.Resource{
				OriginalURL:   item.OriginalURL,
				ShortURL:      encoder.EncodeURL(item.OriginalURL, salt),
				CorrelationID: item.CorrelationID,
			}
		}
		created, err := s.store.CreateBatch(ctx, resources)
		if err != nil {
			return nil, err
		}
		for i := range created {
			created[i].ShortURL = s.baseURL + "/" + created[i].ShortURL
		}
		return created, nil
	})
}
