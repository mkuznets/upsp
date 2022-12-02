package store

import (
	"context"
	"fmt"
	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type dbContextKey string

type Store interface {
	querier(ctx context.Context) pgxtype.Querier

	Payments() Payments
	Tx(ctx context.Context, op func(context.Context) error) error
}

type storeImpl struct {
	pool     *pgxpool.Pool
	payments Payments
}

func New(pool *pgxpool.Pool) Store {
	s := &storeImpl{
		pool: pool,
	}
	s.payments = &paymentsImpl{s: s}
	return s
}

func (s *storeImpl) Payments() Payments {
	return s.payments
}

func (s *storeImpl) querier(ctx context.Context) pgxtype.Querier {
	t := ctx.Value(dbContextKey("tx"))
	if t != nil {
		return t.(pgx.Tx)
	}
	return s.pool
}

func (s *storeImpl) tx(ctx context.Context) (pgx.Tx, bool, error) {
	t := ctx.Value(dbContextKey("tx"))
	if t != nil {
		return t.(pgx.Tx), false, nil
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("could not start transaction: %w", err)
	}
	return tx, true, nil
}

func (s *storeImpl) Tx(ctx context.Context, op func(context.Context) error) error {
	tx, isInner, err := s.tx(ctx)
	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, dbContextKey("tx"), tx)

	if isInner {
		defer func(tx pgx.Tx, ctx context.Context) {
			_ = tx.Rollback(ctx)
		}(tx, ctx)
	}

	if err := op(ctx); err != nil {
		return err
	}

	if isInner {
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("could not commit transaction: %v", err)
		}
	}

	return nil
}
