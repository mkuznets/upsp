package acquirer

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// paymentStore is an interface to create, retrieve, and update payments.
type paymentStore interface {
	// CreateOrGet creates a new payment or returns an existing one with the same ID.
	CreateOrGet(*Payment) (*Payment, error)
	// Get retrieves a payment by ID.
	Get(id PaymentId) (*Payment, error)
	// List retrieves a list of payments by state.
	List(state PaymentState) ([]*Payment, error)
	// Update updates a payment using the given lambda function.
	Update(id PaymentId, version string, fn func(*Payment) error) (*Payment, error)
}

// paymentStoreImpl implements an in-memory thread-safe payment store.
type paymentStoreImpl struct {
	db map[PaymentId]*Payment
	l  *sync.Mutex
}

func newPaymentStore() paymentStore {
	return &paymentStoreImpl{
		db: make(map[PaymentId]*Payment),
		l:  &sync.Mutex{},
	}
}

func (s *paymentStoreImpl) lock(fn func(map[PaymentId]*Payment) error) error {
	s.l.Lock()
	defer s.l.Unlock()
	return fn(s.db)
}

func (s *paymentStoreImpl) CreateOrGet(payment *Payment) (p *Payment, err error) {
	err = s.lock(func(store map[PaymentId]*Payment) error {
		if v, exists := store[payment.Id]; exists {
			p = v
			return nil
		}
		paymentCopy := *payment
		paymentCopy.UpdatedAt = time.Now()
		store[payment.Id] = &paymentCopy
		p = &paymentCopy
		return nil
	})
	return
}

func (s *paymentStoreImpl) Get(id PaymentId) (*Payment, error) {
	if payment, ok := s.db[id]; ok {
		return payment, nil
	}
	return nil, fmt.Errorf("payment not found: %s", id)
}

func (s *paymentStoreImpl) List(state PaymentState) ([]*Payment, error) {
	var payments []*Payment

	err := s.lock(func(store map[PaymentId]*Payment) error {
		for _, payment := range s.db {
			if payment.State() == state {
				payments = append(payments, payment)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return payments, nil
}

func (s *paymentStoreImpl) Update(id PaymentId, version string, fn func(*Payment) error) (*Payment, error) {
	payment := new(Payment)

	err := s.lock(func(store map[PaymentId]*Payment) error {
		if v, ok := store[id]; ok {
			*payment = *v
		} else {
			return fmt.Errorf("payment not found: %s", id)
		}

		if payment.Version != version {
			return fmt.Errorf("version mismatch: %s != %s", payment.Version, version)
		}

		if err := fn(payment); err != nil {
			return err
		}
		payment.Version = uuid.NewString()
		payment.UpdatedAt = time.Now()
		store[id] = payment

		return nil
	})
	if err != nil {
		return nil, err
	}

	return payment, nil
}
