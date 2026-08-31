package repository

import (
	"errors"
	"ya_url_shortener/internal/model"

	"github.com/google/uuid"
)

type MemoryStorage map[uuid.UUID]model.Resource
type LookupStorage map[string]uuid.UUID // в отсутствие БД лукап для поиска ресурсов по коротким юрлам
type Store struct {
	store  MemoryStorage
	lookup LookupStorage
}

var ErrNotFound = errors.New("not found")

func NewStore() *Store {
	return &Store{store: make(MemoryStorage), lookup: make(LookupStorage)}
}

func (s *Store) CreateResource(r model.Resource) model.Resource {
	s.store[r.ID] = r
	s.lookup[r.Shortened] = r.ID
	return r
}

func (s *Store) GetResourceByID(ID uuid.UUID) (model.Resource, error) {
	item, exists := s.store[ID]
	if !exists {
		return model.Resource{}, ErrNotFound
	}
	return item, nil
}

func (s *Store) GetResourceByURL(shortenedURL string) (model.Resource, error) {
	id, exists := s.lookup[shortenedURL]
	if !exists {
		return model.Resource{}, ErrNotFound
	}
	return s.GetResourceByID(id)
}
