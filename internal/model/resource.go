package model

import "github.com/google/uuid"

type Resource struct {
	ID        uuid.UUID
	Address   string
	Shortened string
}
