package repository

import (
	"errors"
	"sync/atomic"
	"ya_url_shortener/internal/model"
)

type MemoryStorage map[int32]model.Resource
type LookupStorage map[string]int32 // в отсутствие БД лукап для поиска ресурсов по коротким юрлам
type Store struct {
	store  MemoryStorage
	lookup LookupStorage
}

var (
	identifier  atomic.Int32
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("integrity error")
)

func NewStore() *Store {
	return &Store{store: make(MemoryStorage), lookup: make(LookupStorage)}
}

func (s *Store) CreateResource(r model.Resource) (model.Resource, error) {
	_, exists := s.lookup[r.Shortened]
	if exists {
		return model.Resource{}, ErrConflict
	}
	if r.ID == 0 {
		r.ID = identifier.Add(1)
	}
	s.store[r.ID] = r
	s.lookup[r.Shortened] = r.ID
	return r, nil
}

func (s *Store) GetResourceByID(ID int32) (model.Resource, error) {
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
