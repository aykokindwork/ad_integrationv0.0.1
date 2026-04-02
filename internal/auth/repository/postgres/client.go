package postgres

import (
	"ad_integration/config"
	"ad_integration/core/apperr"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DbConn struct {
	Pool   *pgxpool.Pool
	config config.DBConfig
}

func NewConnection(ctx context.Context, cfg config.DBConfig) (*DbConn, error) {
	pool, err := pgxpool.New(ctx, cfg.Address)
	if err != nil {
		return nil, apperr.ErrDBUnexpected.WithErr(err)
	}

	return &DbConn{
		Pool:   pool,
		config: cfg,
	}, nil

}
