package store

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type dbContextKey string

type Tx pgx.Tx

type Db interface {
	Tx(ctx context.Context, op func(tx Tx) error) error
}

type dbImpl struct {
	pool *pgxpool.Pool
}

func NewDb(db *pgxpool.Pool) Db {
	return &dbImpl{
		pool: db,
	}
}

func (db *dbImpl) tx(ctx context.Context) (Tx, error) {
	t := ctx.Value(dbContextKey("tx"))
	if t != nil {
		return t.(Tx), nil
	}
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start transaction: %w", err)
	}
	return tx, nil
}

func (db *dbImpl) Tx(ctx context.Context, op func(tx Tx) error) error {
	tx, err := db.tx(ctx)
	if err != nil {
		return err
	}

	defer func(tx Tx, ctx context.Context) {
		_ = tx.Rollback(ctx)
	}(tx, ctx)

	if err := op(tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}
	return nil
}
