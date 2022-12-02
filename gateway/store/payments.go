package store

import (
	"context"
	"github.com/pkg/errors"
	"mkuznets.com/go/gateway/gateway/models"
	"time"
)

type Payments interface {
	Create(ctx context.Context, payment *models.Payment) (string, error)
	Get(ctx context.Context, id string) (*models.Payment, error)
	UpdateState(ctx context.Context, id, version, status string) error
}

type paymentsImpl struct {
	db Db
}

func (s *paymentsImpl) Create(ctx context.Context, payment *models.Payment) (string, error) {
	var id string

	err := s.db.Tx(ctx, func(tx Tx) error {
		err := tx.QueryRow(ctx, `
		INSERT INTO payments (id, merchant_id, amount, currency, card_number, expiry_date, card_holder, cvv, state, "version", created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id;
		`,
			payment.Id,
			payment.MerchantId,
			payment.Amount,
			payment.Currency,
			payment.CardNumber,
			payment.ExpiryDate,
			payment.CardHolder,
			payment.Cvv,
			payment.State,
			payment.Version,
			time.Now(),
			time.Now(),
		).Scan(&id)
		return err
	})

	return id, errors.WithStack(err)
}

func (s *paymentsImpl) Get(ctx context.Context, id string) (*models.Payment, error) {
	var payment models.Payment
	err := s.db.Tx(ctx, func(tx Tx) error {
		err := tx.QueryRow(ctx, `
		SELECT id, merchant_id, amount, currency, card_number, expiry_date, card_holder, cvv, state, version, created_at, updated_at
		FROM payments
		WHERE id = $1;
		`, id).Scan(&payment.Id,
			&payment.MerchantId,
			&payment.Amount,
			&payment.Currency,
			&payment.CardNumber,
			&payment.ExpiryDate,
			&payment.CardHolder,
			&payment.Cvv,
			&payment.State,
			&payment.Version,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		)
		return err
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &payment, nil
}

func (s *paymentsImpl) UpdateState(ctx context.Context, id, version, status string) error {
	return nil
}
