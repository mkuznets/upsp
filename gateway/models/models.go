package models

import (
	acq "mkuznets.com/go/upsp/acquirer"
	"time"
)

type PaymentState string

const (
	PaymentStateProcessing     PaymentState = "processing"
	PaymentStateActionRequired PaymentState = "action_required"
	PaymentStateActionPaid     PaymentState = "paid"

	PaymentStateCancelled PaymentState = "cancelled"
	PaymentStateRefunded  PaymentState = "refunded"
	PaymentStateRejected  PaymentState = "rejected"
)

type Payment struct {
	Id       string
	Amount   int64
	Currency string
	State    PaymentState

	CardNumber string
	CardHolder string
	ExpiryDate string
	Cvv        string

	AcquiringId      string
	AcquiringState   string
	AcquiringVersion string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// SyncState syncs payment state with acquiring state
func (p *Payment) SyncState() {
	switch p.AcquiringState {
	case string(acq.PaymentStateNew), string(acq.PaymentStateAuthorising):
		p.State = PaymentStateProcessing

	case string(acq.PaymentState3dSecureRequired):
		p.State = PaymentStateActionRequired

	case string(acq.PaymentStateConfirmed):
		p.State = PaymentStateActionPaid

	case string(acq.PaymentStateReversed), string(acq.PaymentStateCancelled):
		p.State = PaymentStateCancelled
	case string(acq.PaymentStateRefunded):
		p.State = PaymentStateRefunded
	case string(acq.PaymentStateRejected):
		p.State = PaymentStateRejected
	}
}
