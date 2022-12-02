package models

import "time"

type PaymentState string

const (
	emptyState                   PaymentState = ""
	PaymentStateNew              PaymentState = "new"
	PaymentStateAuthorising      PaymentState = "authorising"
	PaymentState3dSecureRequired PaymentState = "3d_secure_required"
	PaymentStateAuthorised       PaymentState = "authorised"
	PaymentStateConfirmed        PaymentState = "confirmed"

	PaymentStateCancelled PaymentState = "cancelled"
	PaymentStateReversed  PaymentState = "reversed"
	PaymentStateRefunded  PaymentState = "refunded"
	PaymentStateRejected  PaymentState = "rejected"
)

type Payment struct {
	Id         string
	MerchantId string
	Amount     int64
	Currency   string
	State      PaymentState
	Version    string

	CardNumber string
	CardHolder string
	ExpiryDate string
	Cvv        string

	CreatedAt time.Time
	UpdatedAt time.Time
}
