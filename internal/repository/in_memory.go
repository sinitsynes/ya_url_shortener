package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"sync"
	"sync/atomic"

	"ya_url_shortener/internal/model"
)

type memoryStorage map[int32]model.Resource
type lookupStorage map[string]int32 // в отсутствие БД лукап для поиска ресурсов по коротким юрлам
type InMemoryStorage struct {
	identifier atomic.Int32
	store      memoryStorage
	lookup     lookupStorage
	mu         sync.Mutex
}
type InMemoryStore struct {
	InMemoryStorage
	FileStorage
}

func NewStore(storageFileName string) (*InMemoryStore, error) {
	file, err := os.OpenFile(storageFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return nil, ErrCantReadStorage
	}
	s := &InMemoryStore{
		InMemoryStorage: InMemoryStorage{
			store:  make(memoryStorage),
			lookup: make(lookupStorage),
		},
		FileStorage: NewFileStorage(file),
	}
	if backfillErr := s.backfillStore(s.storage); backfillErr != nil {
		return s, backfillErr
	}
	return s, nil
}

func (s *InMemoryStore) backfillStore(f *os.File) error {
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			var r model.Resource
			if err := json.Unmarshal([]byte(line), &r); err != nil {
				return err
			}
			s.store[r.ID] = r
			s.lookup[r.Shortened] = r.ID
			s.identifier.Store(r.ID)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (s *InMemoryStore) CreateResource(r model.Resource) (model.Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.lookup[r.Shortened]
	if exists {
		return model.Resource{}, ErrConflict
	}
	if r.ID == 0 {
		r.ID = s.identifier.Add(1)
	}
	s.store[r.ID] = r
	s.lookup[r.Shortened] = r.ID
	if err := s.AppendToFile(r); err != nil {
		return model.Resource{}, err
	}
	return r, nil
}

func (s *InMemoryStore) getResource(id int32) (model.Resource, error) {
	item, exists := s.store[id]
	if !exists {
		return model.Resource{}, ErrNotFound
	}
	return item, nil
}

func (s *InMemoryStore) GetResourceByID(id int32) (model.Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.getResource(id)
}

func (s *InMemoryStore) GetResourceByURL(shortenedURL string) (model.Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, exists := s.lookup[shortenedURL]
	if !exists {
		return model.Resource{}, ErrNotFound
	}
	return s.getResource(id)
}

// Ping нужен только для совместимости интерфейса с репозиторием на Postgres.
func (s *InMemoryStore) Ping(_ context.Context) error {
	return nil
}
