package repository

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"sync/atomic"

	"ya_url_shortener/internal/model"
)

type MemoryStorage map[int32]model.Resource
type LookupStorage map[string]int32 // в отсутствие БД лукап для поиска ресурсов по коротким юрлам
type Store struct {
	identifier atomic.Int32
	store      MemoryStorage
	lookup     LookupStorage
	storage    *os.File
}

var (
	ErrNotFound        = errors.New("not found")
	ErrConflict        = errors.New("integrity error")
	ErrCantReadStorage = errors.New("can't read storage file")
)

func NewStore(storageFileName string) (*Store, error) {
	storageFile, err := os.OpenFile(storageFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return nil, ErrCantReadStorage
	}
	s := &Store{
		store:   make(MemoryStorage),
		lookup:  make(LookupStorage),
		storage: storageFile,
	}
	if backfillErr := s.backfillStore(s.storage); backfillErr != nil {
		return s, backfillErr
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

func (s *Store) appendToFile(r model.Resource) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	_, err = s.storage.Write(append(data, '\n'))
	return err
}

func (s *Store) CreateResource(r model.Resource) (model.Resource, error) {
	_, exists := s.lookup[r.Shortened]
	if exists {
		return model.Resource{}, ErrConflict
	}
	if r.ID == 0 {
		r.ID = s.identifier.Add(1)
	}
	s.store[r.ID] = r
	s.lookup[r.Shortened] = r.ID
	if err := s.appendToFile(r); err != nil {
		return model.Resource{}, err
	}
	return r, nil
}

func (s *Store) GetResourceByID(id int32) (model.Resource, error) {
	item, exists := s.store[id]
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

func (s *Store) Close() error {
	if s.storage == nil {
		return nil
	}
	return s.storage.Close()
}
