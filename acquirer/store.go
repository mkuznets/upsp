package acquirer

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

//go:generate moq -out store_mock_test.go . Store

// Store is an interface to create, retrieve, and update payments.
type Store interface {
	// CreateOrGet creates a new payment or returns an existing one with the same ID.
	CreateOrGet(*Payment) (*Payment, error)
	// Get retrieves a payment by ID.
	Get(id PaymentId) (*Payment, error)
	// List retrieves a list of payments by state.
	List(state PaymentState) ([]*Payment, error)
	// Update updates a payment using the given lambda function.
	Update(id PaymentId, version string, fn func(*Payment) error) (*Payment, error)
}

// storeImpl implements an in-memory thread-safe payment store.
type storeImpl struct {
	db map[PaymentId]*Payment
	l  *sync.Mutex
}

func NewStore() Store {
	return &storeImpl{
		db: make(map[PaymentId]*Payment),
		l:  &sync.Mutex{},
	}
}

func (s *storeImpl) lock(fn func(map[PaymentId]*Payment) error) error {
	s.l.Lock()
	defer s.l.Unlock()
	return fn(s.db)
}

func (s *storeImpl) CreateOrGet(payment *Payment) (p *Payment, err error) {
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

func (s *storeImpl) Get(id PaymentId) (*Payment, error) {
	if payment, ok := s.db[id]; ok {
		return payment, nil
	}
	return nil, fmt.Errorf("payment not found: %s", id)
}

func (s *storeImpl) List(state PaymentState) ([]*Payment, error) {
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

func (s *storeImpl) Update(id PaymentId, version string, fn func(*Payment) error) (*Payment, error) {
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
