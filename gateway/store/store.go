package store

type Store interface {
	Db() Db
	Payments() Payments
}

type storeImpl struct {
	db       Db
	payments Payments
}

func New(db Db) Store {
	return &storeImpl{
		db:       db,
		payments: &paymentsImpl{db: db},
	}
}

func (s *storeImpl) Payments() Payments {
	return s.payments
}

func (s *storeImpl) Db() Db {
	return s.db
}
