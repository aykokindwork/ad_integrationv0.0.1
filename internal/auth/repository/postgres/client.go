package postgres

import (
	"ad_integration/config"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type DbConn struct {
	Conn   *pgx.Conn
	config config.DBConfig
}

func NewConnection(ctx context.Context, cfg config.DBConfig) (*DbConn, error) {
	conn, err := pgx.Connect(ctx, cfg.Address)
	if err != nil {
		return nil, fmt.Errorf("fail to connect DB: %w", err)
	}

	return &DbConn{
		Conn:   conn,
		config: cfg,
	}, nil

}
