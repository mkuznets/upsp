package store

import (
	"context"
	"mkuznets.com/go/upsp/gateway/models"
	"time"
)

type Payments interface {
	Create(ctx context.Context, payment *models.Payment) (string, error)
	Get(ctx context.Context, id string) (*models.Payment, error)
	ListAll(ctx context.Context) ([]string, error)
	Update(ctx context.Context, id string, op func(payment *models.Payment) error) error
}

type paymentsImpl struct {
	s Store
}

func (p *paymentsImpl) Create(ctx context.Context, payment *models.Payment) (string, error) {
	var id string

	err := p.s.querier(ctx).QueryRow(ctx, `
		INSERT INTO payments (id, amount, currency, card_number, expiry_date, card_holder, cvv, state, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id;
		`,
		payment.Id,
		payment.Amount,
		payment.Currency,
		payment.CardNumber,
		payment.ExpiryDate,
		payment.CardHolder,
		payment.Cvv,
		payment.State,
		time.Now().UTC(),
		time.Now().UTC(),
	).Scan(&id)

	return id, err
}

func (p *paymentsImpl) Get(ctx context.Context, id string) (*models.Payment, error) {
	var payment models.Payment
	err := p.s.querier(ctx).QueryRow(ctx, `
		SELECT id, amount, currency, card_number, expiry_date, card_holder, cvv, state, created_at, updated_at, acquiring_id, acquiring_state, acquiring_version
		FROM payments
		WHERE id = $1;
		`, id).Scan(&payment.Id,
		&payment.Amount,
		&payment.Currency,
		&payment.CardNumber,
		&payment.ExpiryDate,
		&payment.CardHolder,
		&payment.Cvv,
		&payment.State,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.AcquiringId,
		&payment.AcquiringState,
		&payment.AcquiringVersion,
	)
	return &payment, err
}

func (p *paymentsImpl) ListAll(ctx context.Context) ([]string, error) {
	rows, err := p.s.querier(ctx).Query(ctx, `SELECT id FROM payments;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (p *paymentsImpl) Update(ctx context.Context, id string, op func(payment *models.Payment) error) error {
	payment, err := p.Get(ctx, id)
	if err != nil {
		return err
	}
	if err = op(payment); err != nil {
		return err
	}

	_, err = p.s.querier(ctx).Exec(ctx, `
		UPDATE payments
		SET amount = $2,
			currency = $3,
			card_number = $4,
			expiry_date = $5,
			card_holder = $6,
			cvv = $7,
			state = $8,
			acquiring_id = $9,
			acquiring_state = $10,
			acquiring_version = $11,
			updated_at = $12
		WHERE id = $1;
		`,
		payment.Id,
		payment.Amount,
		payment.Currency,
		payment.CardNumber,
		payment.ExpiryDate,
		payment.CardHolder,
		payment.Cvv,
		payment.State,
		payment.AcquiringId,
		payment.AcquiringState,
		payment.AcquiringVersion,
		time.Now().UTC(),
	)
	if err != nil {
		return err
	}

	return nil
}
