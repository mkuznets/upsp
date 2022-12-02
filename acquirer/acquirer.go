package acquirer

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	expected3dsResponse = "123456"
)

type Acquirer interface {
	GetPayment(PaymentId) (*PaymentResource, error)
	CreatePayment(*CreatePaymentRequest) (*CreatePaymentResponse, error)
	AuthorisePayment(id PaymentId, version string, req *AuthorisePaymentRequest) (*AuthorisePaymentResponse, error)
	Submit3dSecure(id PaymentId, version string, req *Submit3dSecureRequest) (*Submit3dSecureResponse, error)
	ConfirmPayment(id PaymentId, version string) (*ConfirmPaymentResponse, error)
	CancelPayment(id PaymentId, version string) (*CancelPaymentResponse, error)
}

type acquirerImpl struct {
	s paymentStore
}

func New() Acquirer {
	a := &acquirerImpl{
		s: newPaymentStore(),
	}
	go a.asyncRefunder()
	go a.asyncTimeouter()

	return a
}

// GetPayment gets a payment.
func (a *acquirerImpl) GetPayment(id PaymentId) (*PaymentResource, error) {
	p, err := a.s.Get(id)
	if err != nil {
		return nil, err
	}

	return &PaymentResource{
		Id:      p.Id,
		State:   p.State(),
		Version: p.Version,
	}, nil
}

// CreatePayment creates a new payment.
func (a *acquirerImpl) CreatePayment(req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	p := &Payment{
		Id:       req.Id,
		Version:  uuid.NewString(),
		Amount:   req.Amount,
		Currency: req.Currency,
	}
	_ = p.SetState(PaymentStateNew)

	m, err := a.s.CreateOrGet(p)
	if err != nil {
		return nil, err
	}

	return &CreatePaymentResponse{
		PaymentResource: PaymentResource{
			Id:      m.Id,
			State:   m.State(),
			Version: m.Version,
		},
	}, nil
}

// AuthorisePayment authorises a payment.
func (a *acquirerImpl) AuthorisePayment(id PaymentId, version string, req *AuthorisePaymentRequest) (*AuthorisePaymentResponse, error) {
	var authUrl string
	p, err := a.s.Update(id, version, func(m *Payment) error {
		if m.State() != PaymentStateNew {
			return fmt.Errorf("payment %s is not in new", m.Id)
		}

		m.CardNumber = req.CardNumber
		m.ExpiryDate = req.ExpiryDate
		m.CardHolder = req.CardHolder
		m.Cvv = req.Cvv

		if is3dSecureRequired(m.CardNumber) {
			if err := m.SetState(PaymentState3dSecureRequired); err != nil {
				return err
			}
			m.Expected3dsResponse = expected3dsResponse
			authUrl = fmt.Sprintf("https://example.com/bank/%s", m.Id)
		} else {
			if err := m.SetState(PaymentStateAuthorising); err != nil {
				return err
			}
			if err := authoriseOrReject(m); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &AuthorisePaymentResponse{
		Payment: PaymentResource{
			Id:      p.Id,
			State:   p.State(),
			Version: p.Version,
		},
		AuthUrl: authUrl,
	}, nil
}

// Submit3dSecure submits a 3d secure response.
func (a *acquirerImpl) Submit3dSecure(id PaymentId, version string, req *Submit3dSecureRequest) (*Submit3dSecureResponse, error) {
	p, err := a.s.Update(id, version, func(m *Payment) error {
		if m.State() != PaymentState3dSecureRequired {
			return fmt.Errorf("payment %s is not in 3d_secure_required", m.Id)
		}

		if m.Expected3dsResponse != req.Token {
			if err := m.SetState(PaymentStateRejected); err != nil {
				return err
			}
		} else {
			if err := m.SetState(PaymentStateAuthorising); err != nil {
				return err
			}
			if err := authoriseOrReject(m); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Submit3dSecureResponse{
		Payment: PaymentResource{
			Id:      p.Id,
			State:   p.State(),
			Version: p.Version,
		},
	}, nil
}

// ConfirmPayment confirms a payment.
func (a *acquirerImpl) ConfirmPayment(id PaymentId, version string) (*ConfirmPaymentResponse, error) {
	p, err := a.s.Update(id, version, func(m *Payment) error {
		return m.SetState(PaymentStateConfirmed)
	})
	if err != nil {
		return nil, err
	}

	return &ConfirmPaymentResponse{
		Payment: PaymentResource{
			Id:      p.Id,
			State:   p.State(),
			Version: p.Version,
		},
	}, nil
}

// CancelPayment cancels a payment.
func (a *acquirerImpl) CancelPayment(id PaymentId, version string) (*CancelPaymentResponse, error) {
	p, err := a.s.Update(id, version, func(m *Payment) error {
		var newState PaymentState

		switch m.State() {
		case PaymentStateNew:
			newState = PaymentStateCancelled
		case PaymentStateAuthorised:
			newState = PaymentStateReversed
		case PaymentStateConfirmed:
			newState = PaymentStateRefunded
		case PaymentState3dSecureRequired:
			newState = PaymentStateRejected
		default:
			return fmt.Errorf("cannot cancel payment in state %s", m.State())
		}

		if err := m.SetState(newState); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &CancelPaymentResponse{
		Payment: PaymentResource{
			Id:      p.Id,
			State:   p.State(),
			Version: p.Version,
		},
	}, nil
}

func authoriseOrReject(p *Payment) error {
	if isSuccess(p.CardNumber) {
		return p.SetState(PaymentStateAuthorised)
	} else {
		return p.SetState(PaymentStateRejected)
	}
}
