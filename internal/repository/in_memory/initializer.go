package inmemory

import (
	filestorage "ya_url_shortener/internal/repository/file_storage"
)

func NewWithFileStorage(filePath string) (*Store, error) {
	fs, err := filestorage.NewFileStorage(filePath)
	if err != nil {
		return &Store{}, err
	}
	r, err := NewStore(fs)
	if err != nil {
		_ = fs.Close()
		return &Store{}, err
	}
	return r, nil
}

func NewInMemoryStore() (*Store, error) {
	r, err := NewStore(nil)
	if err != nil {
		return &Store{}, err
	}
	return r, nil
}
