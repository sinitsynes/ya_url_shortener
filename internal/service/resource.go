package service

import (
	"ya_url_shortener/internal/model"
	"ya_url_shortener/pkg/encoder"

	"github.com/google/uuid"
)

type Repository interface {
	CreateResource(addr string) model.Resource
	GetResourceByID(uuid.UUID) (model.Resource, error)
	UpdateResource(uuid.UUID, model.Resource) (model.Resource, error)
	GetResourceByURL(string) (model.Resource, error)
	SaveToLookup(model.Resource)
}

type Controller struct {
	store Repository
}

func NewResourceController(repository Repository) *Controller {
	return &Controller{store: repository}
}

func (s *Controller) CreateResource(URL string) (model.Resource, error) {
	newResource := s.store.CreateResource(URL)                      // создаем ресурс, получаем идентификатор
	shortenedUrl, err := encoder.EncodeUUIDToString(newResource.ID) //создаем короткий юрл, исходя из хешируемой соли
	if err != nil {
		return model.Resource{}, err
	}
	newResource.Shortened = shortenedUrl
	resource, err := s.store.UpdateResource(newResource.ID, newResource) // обновляем ресурс
	if err != nil {
		return model.Resource{}, err
	}
	s.store.SaveToLookup(resource)
	return resource, nil
}

func (s *Controller) GetResource(shortenedURL string) (model.Resource, error) {
	r, err := s.store.GetResourceByURL(shortenedURL)
	if err != nil {
		return model.Resource{}, err
	}
	return r, nil
}
