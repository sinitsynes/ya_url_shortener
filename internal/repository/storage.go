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

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("integrity error")
)

func NewStore() *Store {
	return &Store{store: make(MemoryStorage), lookup: make(LookupStorage)}
}

func (s *Store) CreateResource(addr string) model.Resource {
	r := model.Resource{ID: uuid.New(), Address: addr}
	s.store[r.ID] = r
	return r
}

func (s *Store) GetResourceByID(ID uuid.UUID) (model.Resource, error) {
	item, exists := s.store[ID]
	if !exists {
		return model.Resource{}, ErrNotFound
	}

	return item, nil
}

func (s *Store) UpdateResource(ID uuid.UUID, toUpdate model.Resource) (model.Resource, error) {
	r, err := s.GetResourceByID(ID)
	if err != nil {
		return r, err
	}
	s.store[r.ID] = toUpdate
	return toUpdate, nil
}

func (s *Store) SaveToLookup(r model.Resource) {
	s.lookup[r.Shortened] = r.ID
}

func (s *Store) GetResourceByURL(shortenedURL string) (model.Resource, error) {
	id, exists := s.lookup[shortenedURL]
	if !exists {
		return model.Resource{}, ErrNotFound
	}
	return s.GetResourceByID(id)
}
