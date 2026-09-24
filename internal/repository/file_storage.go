package repository

import (
	"encoding/json"
	"os"

	"ya_url_shortener/internal/model"
)

type FileStorage struct {
	storage *os.File
}

func NewFileStorage(f *os.File) FileStorage {
	return FileStorage{storage: f}
}

func (s *FileStorage) AppendToFile(r model.Resource) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	_, err = s.storage.Write(append(data, '\n'))
	return err
}

func (s *FileStorage) Close() error {
	if s.storage == nil {
		return nil
	}
	return s.storage.Close()
}
