package filestorage

import (
	"encoding/json"
	"os"

	"ya_url_shortener/internal/model"
	"ya_url_shortener/internal/repository"
)

type FileStorage struct {
	Storage *os.File
}

func NewFileStorage(fileName string) (*FileStorage, error) {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return &FileStorage{}, repository.ErrCantReadStorage
	}
	return &FileStorage{Storage: file}, nil
}

func (s *FileStorage) AppendToFile(r model.Resource) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	_, err = s.Storage.Write(append(data, '\n'))
	return err
}

func (s *FileStorage) Close() error {
	if s.Storage == nil {
		return nil
	}
	return s.Storage.Close()
}
