package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const ConnectionTimeout time.Duration = 1 * time.Second

func InitDB(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(ctx, ConnectionTimeout)
	defer cancel()
	return pgxpool.New(ctx, connString)
}
