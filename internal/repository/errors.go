package repository

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("integrity error")
	ErrOriginalURLConflict = errors.New("original url already exists")
	ErrCantReadStorage     = errors.New("can't read storage file")
	ErrNotImplemented      = errors.New("not implemented")
)

type StorageError struct {
	err     error
	message string
}

func NewRepositoryError(err error, text string) *StorageError {
	return &StorageError{err: err, message: text}
}

func (e *StorageError) Error() string {
	return e.message
}

func (e *StorageError) Unwrap() error {
	return e.err
}

func ErrOriginalURLExists(message string) *StorageError {
	return &StorageError{err: ErrOriginalURLConflict, message: message}
}
