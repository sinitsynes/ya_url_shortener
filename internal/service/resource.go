package service

import (
	"ya_url_shortener/internal/model"
	"ya_url_shortener/internal/repository"
)

type Storager interface {
	CreateResource(addr string) model.Resource
	GetResource(id int32) (model.Resource, error)
	UpdateResource(model.Resource) (model.Resource, error)
}

type Controller struct {
	store Storager
}

func NewResourceController() *Controller {
	return &Controller{store: repository.NewStore()}
}

func (s *Controller) CreateResource(url string) model.Resource {
	newResource := s.store.CreateResource(url)        // создаем ресурс, получаем идентификатор
	shortenedUrl := encodeUrl(newResource.Identifier) //создаем короткий юрл, исходя из айди
	newResource.Shortened = shortenedUrl
	s.store.UpdateResource(newResource) // обновляем ресурс
	return newResource
}

func (s *Controller) GetResource(id int32) (model.Resource, error) {
	r, err := s.store.GetResource(id)
	if err != nil {
		return r, err
	}
	return r, nil
}
