package db

import (
	"context"
	"errors"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type UnavailablePinger struct{}

func (UnavailablePinger) Ping(context.Context) error {
	return errors.New("database is not configured")
}
