package acquirer

import (
	"fmt"

	"github.com/google/uuid"
	"mkuznets.com/go/gateway/acquirer/models"
	"mkuznets.com/go/gateway/acquirer/store"
)

const (
	expected3dsResponse = "123456"
)

type Acquirer interface {
	GetPayment(models.PaymentId) (*PaymentResource, error)
	CreatePayment(*CreatePaymentRequest) (*CreatePaymentResponse, error)
	AuthorisePayment(id models.PaymentId, version string, req *AuthorisePaymentRequest) (*AuthorisePaymentResponse, error)
	Submit3dSecure(id models.PaymentId, version string, req *Submit3dSecureRequest) (*Submit3dSecureResponse, error)
	ConfirmPayment(id models.PaymentId, version string) (*ConfirmPaymentResponse, error)
	CancelPayment(id models.PaymentId, version string) (*CancelPaymentResponse, error)
}

type acquirerImpl struct {
	s store.PaymentStore
}

func NewAcquirer() (Acquirer, error) {
	s, err := store.NewStore()
	if err != nil {
		return nil, err
	}
	a := &acquirerImpl{s: s}

	return a, nil
}

// GetPayment gets a payment.
func (a *acquirerImpl) GetPayment(id models.PaymentId) (*PaymentResource, error) {
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
	p := &models.Payment{
		Id:       req.Id,
		Version:  uuid.NewString(),
		Amount:   req.Amount,
		Currency: req.Currency,
		HookUrl:  req.HookUrl,
	}
	_ = p.SetState(models.PaymentStateNew)

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
func (a *acquirerImpl) AuthorisePayment(id models.PaymentId, version string, req *AuthorisePaymentRequest) (*AuthorisePaymentResponse, error) {
	var authUrl string
	p, err := a.s.Update(id, version, func(m *models.Payment) error {
		if m.State() != models.PaymentStateNew {
			return fmt.Errorf("payment %s is not in new", m.Id)
		}

		m.CardNumber = req.CardNumber
		m.ExpiryDate = req.ExpiryDate
		m.CardHolder = req.CardHolder
		m.Cvv = req.Cvv

		if m.Is3DSecureRequired() {
			if err := m.SetState(models.PaymentState3dSecureRequired); err != nil {
				return err
			}
			m.Expected3dsResponse = expected3dsResponse
			authUrl = "https://example.com/3ds/" + m.Expected3dsResponse
		} else {
			if err := m.SetState(models.PaymentStateAuthorising); err != nil {
				return err
			}
			if err := m.AuthoriseOrReject(); err != nil {
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
func (a *acquirerImpl) Submit3dSecure(id models.PaymentId, version string, req *Submit3dSecureRequest) (*Submit3dSecureResponse, error) {
	p, err := a.s.Update(id, version, func(m *models.Payment) error {
		if m.State() != models.PaymentState3dSecureRequired {
			return fmt.Errorf("payment %s is not in 3d_secure_required", m.Id)
		}

		if m.Expected3dsResponse != req.Token {
			if err := m.SetState(models.PaymentStateRejected); err != nil {
				return err
			}
		} else {
			if err := m.SetState(models.PaymentStateAuthorising); err != nil {
				return err
			}
			if err := m.AuthoriseOrReject(); err != nil {
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
func (a *acquirerImpl) ConfirmPayment(id models.PaymentId, version string) (*ConfirmPaymentResponse, error) {
	p, err := a.s.Update(id, version, func(m *models.Payment) error {
		return m.SetState(models.PaymentStateConfirmed)
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
func (a *acquirerImpl) CancelPayment(id models.PaymentId, version string) (*CancelPaymentResponse, error) {
	p, err := a.s.Update(id, version, func(m *models.Payment) error {
		var newState models.PaymentState

		switch m.State() {
		case models.PaymentStateNew:
			newState = models.PaymentStateCancelled
		case models.PaymentStateAuthorised:
			newState = models.PaymentStateReversed
		case models.PaymentStateConfirmed:
			newState = models.PaymentStateRefunded
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
