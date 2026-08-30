package model

import "github.com/google/uuid"

type Resource struct {
	Identifier int32
	Salt       uuid.UUID
	Address    string
	Shortened  string
}
