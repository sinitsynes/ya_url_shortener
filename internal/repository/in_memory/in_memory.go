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
			s.lookup[r.ShortURL] = r.ID
			s.identifier.Store(r.ID)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (s *Store) saveToStore(r model.Resource) model.Resource {
	if r.ID == 0 {
		r.ID = s.identifier.Add(1)
	}
	s.store[r.ID] = r
	s.lookup[r.ShortURL] = r.ID
	return r
}

func (s *Store) CreateResource(_ context.Context, r model.Resource) (model.Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.lookup[r.ShortURL]
	if exists {
		return model.Resource{}, repository.ErrConflict
	}
	created := s.saveToStore(r)
	if s.fileStorage != nil {
		if err := s.fileStorage.AppendToFile(created); err != nil {
			return model.Resource{}, err
		}
	}
	return created, nil
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

func (s *Store) GetResourceByShortURL(_ context.Context, shortenedURL string) (model.Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, exists := s.lookup[shortenedURL]
	if !exists {
		return model.Resource{}, repository.ErrNotFound
	}
	return s.getResource(id)
}

func (s *Store) GetResourceByOriginalURL(_ context.Context, originalURL string) (model.Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range s.store {
		if item.OriginalURL == originalURL {
			return item, nil
		}
	}
	return model.Resource{}, repository.ErrNotFound
}

func (s *Store) CreateBatch(_ context.Context, resources []model.Resource) ([]model.Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	output := make([]model.Resource, len(resources))
	// до записи в мапу проверяем, что конфликтов не будет ни с одним из элементов списка
	for _, item := range resources {
		_, exists := s.lookup[item.ShortURL]
		if exists {
			return nil, repository.ErrConflict
		}
	}
	// если ни с кем конфликтов нет, то записываем весь список
	for index, item := range resources {
		created := s.saveToStore(item)
		if s.fileStorage != nil {
			if err := s.fileStorage.AppendToFile(created); err != nil {
				return nil, err
			}
		}
		output[index] = model.Resource{
			ID:            created.ID,
			OriginalURL:   created.OriginalURL,
			CorrelationID: created.CorrelationID,
			ShortURL:      created.ShortURL,
		}
	}
	return output, nil
}
