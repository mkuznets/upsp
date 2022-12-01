package store

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"mkuznets.com/go/gateway/acquirer/models"
)

type PaymentStore interface {
	CreateOrGet(*models.Payment) (*models.Payment, error)
	Get(id models.PaymentId) (*models.Payment, error)
	List(state models.PaymentState) ([]*models.Payment, error)
	Update(id models.PaymentId, version string, fn func(*models.Payment) error) (*models.Payment, error)
}

type paymentStoreImpl struct {
	db map[models.PaymentId]*models.Payment
	l  *sync.Mutex
}

func NewStore() (PaymentStore, error) {
	return &paymentStoreImpl{
		db: make(map[models.PaymentId]*models.Payment),
		l:  &sync.Mutex{},
	}, nil
}

func (s *paymentStoreImpl) lock(fn func(map[models.PaymentId]*models.Payment) error) error {
	s.l.Lock()
	defer s.l.Unlock()
	return fn(s.db)
}

func (s *paymentStoreImpl) CreateOrGet(payment *models.Payment) (p *models.Payment, err error) {
	err = s.lock(func(store map[models.PaymentId]*models.Payment) error {
		if v, exists := store[payment.Id]; exists {
			p = v
			return nil
		}
		store[payment.Id] = payment
		p = payment
		return nil
	})
	return
}

func (s *paymentStoreImpl) Get(id models.PaymentId) (*models.Payment, error) {
	if payment, ok := s.db[id]; ok {
		return payment, nil
	}
	return nil, fmt.Errorf("payment not found: %s", id)
}

func (s *paymentStoreImpl) List(state models.PaymentState) ([]*models.Payment, error) {
	var payments []*models.Payment

	err := s.lock(func(store map[models.PaymentId]*models.Payment) error {
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

func (s *paymentStoreImpl) Update(id models.PaymentId, version string, fn func(*models.Payment) error) (*models.Payment, error) {
	payment := new(models.Payment)

	err := s.lock(func(store map[models.PaymentId]*models.Payment) error {
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
		store[id] = payment
		payment.OnUpdate()

		return nil
	})
	if err != nil {
		return nil, err
	}

	return payment, nil
}
