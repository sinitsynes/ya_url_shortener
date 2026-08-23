package repository

import (
	"errors"
	"sync/atomic"
	"ya_url_shortener/internal/model"
)

type MemoryStorage map[int32]model.Resource

var (
	counter     atomic.Int32
	ErrNotFound = errors.New("not found")
)

func NewStorage() MemoryStorage {
	return make(MemoryStorage)
}

func CreateResource(storage MemoryStorage, addr string) model.Resource {
	id := counter.Add(1)
	r := model.Resource{Identifier: id, Address: addr}
	storage[id] = r
	return r
}

func GetResource(storage MemoryStorage, id int32) (model.Resource, error) {
	item, exists := storage[id]
	if !exists {
		return model.Resource{}, ErrNotFound
	}
	return item, nil
}
