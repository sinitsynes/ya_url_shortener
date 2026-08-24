package repository

import (
	"errors"
	"sync/atomic"
	"ya_url_shortener/internal/model"
)

type MemoryStorage map[int32]model.Resource
type Store struct {
	store MemoryStorage
}

var (
	counter     atomic.Int32
	ErrNotFound = errors.New("not found")
)

func NewStore() *Store {
	return &Store{make(MemoryStorage)}
}

func (s *Store) CreateResource(addr string) model.Resource {
	id := counter.Add(1)
	r := model.Resource{Identifier: id, Address: addr}
	s.store[id] = r
	return r
}

func (s *Store) GetResource(id int32) (model.Resource, error) {
	item, exists := s.store[id]
	if !exists {
		return model.Resource{}, ErrNotFound
	}
	return item, nil
}

func (s *Store) UpdateResource(r model.Resource) (model.Resource, error) {
	r, err := s.GetResource(r.Identifier)
	if err != nil {
		return r, err
	}
	s.store[r.Identifier] = r
	return r, nil
}
