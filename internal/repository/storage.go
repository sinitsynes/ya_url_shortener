package repository

import (
	"errors"
	"sync/atomic"
	"ya_url_shortener/internal/model"

	"github.com/google/uuid"
)

type MemoryStorage map[int32]model.Resource
type Store struct {
	store MemoryStorage
}

var (
	counter     atomic.Int32
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("integrity error")
)

func NewStore() *Store {
	return &Store{make(MemoryStorage)}
}

func (s *Store) CreateResource(addr string) model.Resource {
	id := counter.Add(1)
	salt := uuid.New()
	r := model.Resource{Identifier: id, Address: addr, Salt: salt}
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

func (s *Store) UpdateResource(id int32, toUpdate model.Resource) (model.Resource, error) {
	r, err := s.GetResource(id)
	if err != nil {
		return r, err
	}
	s.store[r.Identifier] = toUpdate
	return toUpdate, nil
}
