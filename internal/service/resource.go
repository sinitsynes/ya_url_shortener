package service

import (
	"context"
	"errors"
	"uuid"

	"ya_url_shortener/internal/model"
	"ya_url_shortener/internal/repository"
	"ya_url_shortener/pkg/encoder"
)

const maxCreateAttempts int32 = 5

var ErrMaxRetriesExceeded = errors.New("failed to generate unique shortened url")

type Repository interface {
	CreateResource(context.Context, model.Resource) (model.Resource, error)
	GetResourceByID(context.Context, int32) (model.Resource, error)
	GetResourceByShortURL(context.Context, string) (model.Resource, error)
	GetResourceByOriginalURL(context.Context, string) (model.Resource, error)
	GetResourcesByUserID(context.Context, *uuid.UUID) ([]model.Resource, error)
	CreateBatch(context.Context, []model.Resource) ([]model.Resource, error)
	SoftDeleteResources(context.Context, []model.ToDelete) error
	Close() error
}

type Controller struct {
	baseURL    string
	store      Repository
	deleteCh   chan model.ToDelete
	deleteDone chan struct{}
}

func NewResourceController(baseURL string, repository Repository) *Controller {
	c := &Controller{
		baseURL:    baseURL,
		store:      repository,
		deleteCh:   make(chan model.ToDelete, deleteBatchSize),
		deleteDone: make(chan struct{}),
	}
	go c.deleteWorker()
	return c
}

// withConflictRetry повторяет попытку создания записей,
// пока количество попыток не перевалит за константу.
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

func (c *Controller) CreateResource(
	ctx context.Context,
	originalURL string,
	userID *uuid.UUID,
) (model.Resource, error) {
	existing, err := c.store.GetResourceByOriginalURL(ctx, originalURL)
	if err == nil {
		existing.ShortURL = c.baseURL + "/" + existing.ShortURL
		return existing, repository.ErrOriginalURLConflict
	}
	return withConflictRetry(func(salt int32) (model.Resource, error) {
		created, createErr := c.store.CreateResource(ctx, model.Resource{
			OriginalURL: originalURL,
			ShortURL:    encoder.EncodeURL(originalURL, salt),
			UserID:      userID,
		})
		if createErr != nil {
			return model.Resource{}, createErr
		}
		created.ShortURL = c.baseURL + "/" + created.ShortURL
		return created, nil
	})
}

func (c *Controller) GetResource(ctx context.Context, shortenedURL string) (model.Resource, error) {
	r, err := c.store.GetResourceByShortURL(ctx, shortenedURL)
	if err != nil {
		return model.Resource{}, err
	}
	r.ShortURL = c.baseURL + "/" + r.ShortURL
	return r, nil
}

func (c *Controller) CreateBatch(
	ctx context.Context,
	batch []model.ResourceBatchInput,
	userID *uuid.UUID,
) ([]model.ResourceBatchOutput, error) {
	return withConflictRetry(func(salt int32) ([]model.ResourceBatchOutput, error) {
		resources := make([]model.Resource, len(batch))
		for i, item := range batch {
			resources[i] = model.Resource{
				OriginalURL:   item.OriginalURL,
				ShortURL:      encoder.EncodeURL(item.OriginalURL, salt),
				CorrelationID: &item.CorrelationID,
				UserID:        userID,
			}
		}
		created, err := c.store.CreateBatch(ctx, resources)
		if err != nil {
			return nil, err
		}
		out := make([]model.ResourceBatchOutput, len(created))
		for i, item := range created {
			out[i] = model.ResourceBatchOutput{
				CorrelationID: *item.CorrelationID,
				ShortURL:      c.baseURL + "/" + item.ShortURL,
			}
		}
		return out, nil
	})
}

func (c *Controller) GetResourcesByUser(ctx context.Context, userID *uuid.UUID) ([]model.ResourcesByUser, error) {
	resources, err := c.store.GetResourcesByUserID(ctx, userID)
	out := make([]model.ResourcesByUser, len(resources))
	if err != nil {
		return out, err
	}
	for i, item := range resources {
		out[i] = model.ResourcesByUser{
			OriginalURL: item.OriginalURL,
			ShortURL:    c.baseURL + "/" + item.ShortURL,
		}
	}
	return out, nil
}

// DeleteResources удаляет ресурсы в отдельной очереди:
// - либо по достижению батча в 50 ссылок
// - либо по таймеру: если одна ссылка лежит долго в очереди, то через секунду ее удалим.
func (c *Controller) DeleteResources(ctx context.Context, userID *uuid.UUID, shortURLs []string) error {
	for _, shortURL := range shortURLs {
		select {
		case <-ctx.Done():
			return nil
		case c.deleteCh <- model.ToDelete{UserID: *userID, ShortURL: shortURL}:
		}
	}
	return nil
}
