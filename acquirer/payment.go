package acquirer

import (
	"fmt"
	"time"
)

type (
	PaymentId    string
	PaymentState string
)

const (
	emptyState PaymentState = ""

	// PaymentStateNew is the initial state of a payment.
	PaymentStateNew PaymentState = "new"

	// PaymentStateAuthorising is the state of a payment that is being authorised.
	PaymentStateAuthorising PaymentState = "authorising"

	// PaymentState3dSecureRequired is the state of a payment that is waiting for 3DS action by the customer.
	PaymentState3dSecureRequired PaymentState = "3d_secure_required"

	// PaymentStateAuthorised is the state of a payment that has been successfully authorised but not yet paid.
	PaymentStateAuthorised PaymentState = "authorised"

	// PaymentStateConfirmed is the state of a payment that has been paid.
	PaymentStateConfirmed PaymentState = "confirmed"

	// PaymentStateCancelled is the state of a payment that has been cancelled before authorisation. Final state.
	PaymentStateCancelled PaymentState = "cancelled"

	// PaymentStateReversed is the state of a payment that has been cancelled after authorisation. Final state.
	PaymentStateReversed PaymentState = "reversed"

	// PaymentStateRefunded is the state of a payment that has been refunded after its confirmation. Final state.
	PaymentStateRefunded PaymentState = "refunded"

	// PaymentStateRejected is the state of a payment that failed the authorisation step. Final state.
	PaymentStateRejected PaymentState = "rejected"
)

var validTransactions = map[PaymentState][]PaymentState{
	emptyState:                   {PaymentStateNew},
	PaymentStateNew:              {PaymentStateAuthorising, PaymentState3dSecureRequired, PaymentStateCancelled},
	PaymentStateAuthorising:      {PaymentStateAuthorised, PaymentStateRejected},
	PaymentState3dSecureRequired: {PaymentStateAuthorising, PaymentStateRejected},
	PaymentStateAuthorised:       {PaymentStateConfirmed, PaymentStateReversed},
	PaymentStateConfirmed:        {PaymentStateRefunded},
	PaymentStateRejected:         {},
}

// Payment is a record that represents a stored payment at the acquiring bank.
type Payment struct {
	Id        PaymentId
	state     PaymentState
	prevState PaymentState
	Version   string

	Amount   int64
	Currency string

	CardNumber string
	ExpiryDate string
	CardHolder string
	Cvv        string

	UpdatedAt time.Time

	Expected3dsResponse string
}

// State returns the state of the payment.
func (p *Payment) State() PaymentState {
	return p.state
}

// SetState sets the state of the payment. Returns an error if the transition is invalid.
func (p *Payment) SetState(state PaymentState) error {
	if !isValidTransition(p.state, state) {
		return fmt.Errorf("invalid payment transition: %s -> %s", p.state, state)
	}
	p.prevState = p.state
	p.state = state
	return nil
}

func isValidTransition(from, to PaymentState) bool {
	for _, state := range validTransactions[from] {
		if state == to {
			return true
		}
	}
	return false
}
