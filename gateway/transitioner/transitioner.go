package transitioner

import (
	"context"
	"github.com/google/uuid"
	"log"
	"mkuznets.com/go/upsp/acquirer"
	"mkuznets.com/go/upsp/gateway/models"
	"mkuznets.com/go/upsp/gateway/store"
	"time"
)

// Transitioner is a service that transitions payments through the acquiring process.
// It works both as a synchronous transitioner and as a background worker.
type Transitioner interface {
	Start(ctx context.Context)
	Transition(ctx context.Context, id string) error
}

type transitionerImpl struct {
	s   store.Store
	acq acquirer.Acquirer
}

// New creates a new Transitioner.
func New(s store.Store, acq acquirer.Acquirer) Transitioner {
	t := &transitionerImpl{
		s:   s,
		acq: acq,
	}
	return t
}

// Start initiates a background worker that scans all gateway payments and syncs their status with the acquirer.
func (t *transitionerImpl) Start(ctx context.Context) {
	for {
		ids, err := t.s.Payments().ListAll(ctx)
		if err != nil {
			log.Printf("[ERR] %v", err)
			continue
		}

		for _, id := range ids {
			err := t.Transition(ctx, id)
			if err != nil {
				// Skip the payment since it is likely not persisted in the acquirer.
				continue
			}
		}

		time.Sleep(5 * time.Second)
	}

}

// Transition synchronously transitions a payment of the given ID through the acquiring process.
// The transition stops when the payment reaches one of the final states, PaymentStateActionRequired, or an error occurs.
func (t *transitionerImpl) Transition(ctx context.Context, id string) error {
	return t.s.Tx(ctx, func(ctx context.Context) error {
		for {
			p, err := t.s.Payments().Get(ctx, id)
			if err != nil {
				return err
			}

			switch p.AcquiringState {
			case "":
				if errC := t.initialisePayment(ctx, p); errC != nil {
					return errC
				}

			case string(acquirer.PaymentStateAuthorised):
				if errC := t.confirmPayment(ctx, p); errC != nil {
					return errC
				}

			default:
				return t.syncPayment(ctx, p)
			}
		}
	})
}

func (t *transitionerImpl) initialisePayment(ctx context.Context, payment *models.Payment) error {
	aId := uuid.NewString()

	rCreate, err := t.acq.CreatePayment(&acquirer.CreatePaymentRequest{
		Id:       acquirer.PaymentId(aId),
		Amount:   payment.Amount,
		Currency: payment.Currency,
	})
	if err != nil {
		return err
	}

	rAuth, err := t.acq.AuthorisePayment(rCreate.Id, rCreate.Version, &acquirer.AuthorisePaymentRequest{
		CardNumber: payment.CardNumber,
		ExpiryDate: payment.ExpiryDate,
		CardHolder: payment.CardHolder,
		Cvv:        payment.Cvv,
	})

	return t.s.Payments().Update(ctx, payment.Id, func(py *models.Payment) error {
		py.AcquiringId = aId
		py.AcquiringVersion = rAuth.Payment.Version
		py.AcquiringState = string(rAuth.Payment.State)
		py.SyncState()
		return nil
	})
}

func (t *transitionerImpl) confirmPayment(ctx context.Context, payment *models.Payment) error {
	rConfirm, err := t.acq.ConfirmPayment(acquirer.PaymentId(payment.AcquiringId), payment.AcquiringVersion)
	if err != nil {
		return err
	}

	return t.s.Payments().Update(ctx, payment.Id, func(py *models.Payment) error {
		py.AcquiringVersion = rConfirm.Payment.Version
		py.AcquiringState = string(rConfirm.Payment.State)
		py.SyncState()
		return nil
	})
}

func (t *transitionerImpl) syncPayment(ctx context.Context, payment *models.Payment) error {
	rGet, err := t.acq.GetPayment(acquirer.PaymentId(payment.AcquiringId))
	if err != nil {
		return err
	}
	if payment.AcquiringVersion == rGet.Version {
		return nil
	}

	err = t.s.Payments().Update(ctx, payment.Id, func(py *models.Payment) error {
		py.AcquiringVersion = rGet.Version
		py.AcquiringState = string(rGet.State)
		py.SyncState()
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
