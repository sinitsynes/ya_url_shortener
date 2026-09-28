package service

import (
	"context"
	"errors"

	"ya_url_shortener/internal/model"
	"ya_url_shortener/internal/repository"
	"ya_url_shortener/pkg/encoder"
)

const maxCreateAttempts = 5

var ErrMaxRetriesExceeded = errors.New("failed to generate unique shortened url")

type Repository interface {
	CreateResource(context.Context, model.Resource) (model.Resource, error)
	GetResourceByID(context.Context, int32) (model.Resource, error)
	GetResourceByURL(context.Context, string) (model.Resource, error)
}

type Controller struct {
	store Repository
}

func NewResourceController(repository Repository) *Controller {
	return &Controller{store: repository}
}

func (s *Controller) CreateResource(ctx context.Context, originalURL string) (model.Resource, error) {
	newResource := model.Resource{Address: originalURL}

	for saltCounter := range maxCreateAttempts {
		newResource.Shortened = encoder.EncodeURL(originalURL, int32(saltCounter))

		created, err := s.store.CreateResource(ctx, newResource)
		if err == nil {
			return created, nil
		}
		if !errors.Is(err, repository.ErrConflict) {
			return model.Resource{}, err
		}
	}

	return model.Resource{}, ErrMaxRetriesExceeded
}

func (s *Controller) GetResource(ctx context.Context, shortenedURL string) (model.Resource, error) {
	r, err := s.store.GetResourceByURL(ctx, shortenedURL)
	if err != nil {
		return model.Resource{}, err
	}
	return r, nil
}
