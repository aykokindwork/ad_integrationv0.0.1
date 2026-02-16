package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func Connection(ctx context.Context, address string) (*pgx.Conn, error) {
	return pgx.Connect(ctx, address)
}
