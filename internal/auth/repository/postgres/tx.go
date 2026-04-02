package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionManager struct {
	db *pgxpool.Conn
}

func NewTranscationManager(db *pgxpool.Conn) *TransactionManager {
	return &TransactionManager{
		db: db,
	}
}

type txKey struct{}

type Querier interface {
	Get(ctx context.Context, db Querier, dst interface{}, query string, args ...interface{}) error
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}


func (tm *TransactionManager) W
