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
	store Repository
}

func NewResourceController(repository Repository) *Controller {
	return &Controller{store: repository}
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
		return s.store.CreateResource(ctx, model.Resource{
			OriginalURL: originalURL,
			ShortURL:    encoder.EncodeURL(originalURL, salt),
		})
	})
}

func (s *Controller) GetResource(ctx context.Context, shortenedURL string) (model.Resource, error) {
	r, err := s.store.GetResourceByURL(ctx, shortenedURL)
	if err != nil {
		return model.Resource{}, err
	}
	return r, nil
}

func (s *Controller) CreateBatch(
	ctx context.Context,
	batch []model.ResourceBatchInput,
) ([]model.ResourceBatchOutput, error) {
	return withConflictRetry(func(salt int32) ([]model.ResourceBatchOutput, error) {
		resources := make([]model.Resource, 0, len(batch))
		for _, item := range batch {
			resources = append(resources, model.Resource{
				OriginalURL:   item.OriginalURL,
				ShortURL:      encoder.EncodeURL(item.OriginalURL, salt),
				CorrelationID: item.CorrelationID,
			})
		}
		return s.store.CreateBatch(ctx, resources)
	})
}
