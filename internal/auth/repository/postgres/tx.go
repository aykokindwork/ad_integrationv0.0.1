package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionManager struct {
	db *pgxpool.Pool
}

func NewTranscationManager(db *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{
		db: db,
	}
}

type txKey struct{}

type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
}

func (tm *TransactionManager) WithInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.db.Begin(ctx)
	if err != nil {
		return err
	}

	ctxWithTx := context.WithValue(ctx, txKey{}, tx)

	err = fn(ctxWithTx)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
