package service

import (
	"ya_url_shortener/internal/model"
	"ya_url_shortener/pkg/encoder"
)

type Repository interface {
	CreateResource(addr string) model.Resource
	GetResource(int32) (model.Resource, error)
	UpdateResource(int32, model.Resource) (model.Resource, error)
}

type Controller struct {
	store Repository
}

func NewResourceController(repository Repository) *Controller {
	return &Controller{store: repository}
}

func (s *Controller) CreateResource(url string) (model.Resource, error) {
	newResource := s.store.CreateResource(url)                        // создаем ресурс, получаем идентификатор
	shortenedUrl, err := encoder.EncodeUUIDToString(newResource.Salt) //создаем короткий юрл, исходя из хешируемой соли
	if err != nil {
		return model.Resource{}, err
	}
	newResource.Shortened = shortenedUrl
	resource, err := s.store.UpdateResource(newResource.Identifier, newResource) // обновляем ресурс
	if err != nil {
		return model.Resource{}, err
	}
	return resource, nil
}

func (s *Controller) GetResource(id int32) (model.Resource, error) {
	r, err := s.store.GetResource(id)
	if err != nil {
		return r, err
	}
	return r, nil
}
