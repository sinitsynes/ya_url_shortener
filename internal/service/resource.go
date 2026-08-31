package service

import (
	"ya_url_shortener/internal/model"
	"ya_url_shortener/pkg/encoder"

	"github.com/google/uuid"
)

type Repository interface {
	CreateResource(model.Resource) model.Resource
	GetResourceByID(uuid.UUID) (model.Resource, error)
	GetResourceByURL(string) (model.Resource, error)
}

type Controller struct {
	store Repository
}

func NewResourceController(repository Repository) *Controller {
	return &Controller{store: repository}
}

func (s *Controller) CreateResource(originalUrl string) (model.Resource, error) {
	id := uuid.New()
	shortened, err := encoder.EncodeUUIDToString(id)
	if err != nil {
		return model.Resource{}, err
	}
	resource := model.Resource{
		ID:        id,
		Address:   originalUrl,
		Shortened: shortened,
	}
	newResource := s.store.CreateResource(resource)
	return newResource, nil
}

func (s *Controller) GetResource(shortenedURL string) (model.Resource, error) {
	r, err := s.store.GetResourceByURL(shortenedURL)
	if err != nil {
		return model.Resource{}, err
	}
	return r, nil
}
