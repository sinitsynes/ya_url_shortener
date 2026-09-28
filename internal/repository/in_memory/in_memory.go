package inmemory

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"sync"
	"sync/atomic"

	"ya_url_shortener/internal/model"
	"ya_url_shortener/internal/repository"
	fs "ya_url_shortener/internal/repository/file_storage"
)

type (
	memoryStorage   map[int32]model.Resource
	lookupStorage   map[string]int32 // в отсутствие БД лукап для поиска ресурсов по коротким юрлам
	inMemoryStorage struct {
		identifier atomic.Int32
		store      memoryStorage
		lookup     lookupStorage
		mu         sync.Mutex
	}
	Store struct {
		inMemoryStorage

		fileStorage *fs.FileStorage
	}
)

func NewStore(fileStorage *fs.FileStorage) (*Store, error) {
	s := &Store{
		inMemoryStorage: inMemoryStorage{
			store:  make(memoryStorage),
			lookup: make(lookupStorage),
		}}

	if fileStorage != nil {
		s.fileStorage = fileStorage

		if backfillErr := s.backfillStore(s.fileStorage.Storage); backfillErr != nil {
			return s, backfillErr
		}
	}
	return s, nil
}

func (s *Store) backfillStore(f *os.File) error {
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

func (s *Store) CreateResource(_ context.Context, r model.Resource) (model.Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.lookup[r.Shortened]
	if exists {
		return model.Resource{}, repository.ErrConflict
	}
	if r.ID == 0 {
		r.ID = s.identifier.Add(1)
	}
	s.store[r.ID] = r
	s.lookup[r.Shortened] = r.ID
	if s.fileStorage != nil {
		if err := s.fileStorage.AppendToFile(r); err != nil {
			return model.Resource{}, err
		}
	}
	return r, nil
}

func (s *Store) getResource(id int32) (model.Resource, error) {
	item, exists := s.store[id]
	if !exists {
		return model.Resource{}, repository.ErrNotFound
	}
	return item, nil
}

func (s *Store) GetResourceByID(_ context.Context, id int32) (model.Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.getResource(id)
}

func (s *Store) GetResourceByURL(_ context.Context, shortenedURL string) (model.Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, exists := s.lookup[shortenedURL]
	if !exists {
		return model.Resource{}, repository.ErrNotFound
	}
	return s.getResource(id)
}

// Close закрывает файл, если используется запись в него. Если нет, то это no-op.
func (s *Store) Close() error {
	if s.fileStorage != nil {
		if err := s.fileStorage.Close(); err != nil {
			return err
		}
	}
	return nil
}
