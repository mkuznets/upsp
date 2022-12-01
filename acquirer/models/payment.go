package models

import (
	"fmt"
)

type (
	PaymentId    string
	PaymentState string
)

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

var validTransactions = map[PaymentState][]PaymentState{
	emptyState:                   {PaymentStateNew},
	PaymentStateNew:              {PaymentStateAuthorising, PaymentState3dSecureRequired, PaymentStateCancelled},
	PaymentStateAuthorising:      {PaymentStateAuthorised, PaymentStateRejected},
	PaymentState3dSecureRequired: {PaymentStateAuthorising, PaymentStateRejected},
	PaymentStateAuthorised:       {PaymentStateConfirmed, PaymentStateReversed},
	PaymentStateConfirmed:        {PaymentStateRefunded},
	PaymentStateRejected:         {},
}

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

	Expected3dsResponse string

	HookUrl string
}

// State returns the state of the payment.
func (p *Payment) State() PaymentState {
	return p.state
}

// SetState sets the state of the payment.
func (p *Payment) SetState(state PaymentState) error {
	if !isValidTransition(p.state, state) {
		return fmt.Errorf("invalid payment transition: %s -> %s", p.state, state)
	}
	p.prevState = p.state
	p.state = state
	return nil
}

func (p *Payment) AuthoriseOrReject() error {
	if isAuthorisationSuccessful(p.CardNumber) {
		return p.SetState(PaymentStateAuthorised)
	} else {
		return p.SetState(PaymentStateRejected)
	}
}

func (p *Payment) Is3DSecureRequired() bool {
	return is3dSecureRequired(p.CardNumber)
}

func (p *Payment) OnUpdate() {
	if p.prevState == p.state {
		return
	}
	fmt.Println("payment updated:", p.Id, p.prevState, "->", p.state)
	//http.Post(p.HookUrl, "application/json", nil)
}

func isValidTransition(from, to PaymentState) bool {
	for _, state := range validTransactions[from] {
		if state == to {
			return true
		}
	}
	return false
}
