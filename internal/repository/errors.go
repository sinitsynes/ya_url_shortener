package repository

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrConflict        = errors.New("integrity error")
	ErrCantReadStorage = errors.New("can't read storage file")
	ErrNotImplemented  = errors.New("not implemented")
)
