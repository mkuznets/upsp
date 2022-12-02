package acquirer

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_paymentStoreImpl_CreateOrGet(t *testing.T) {

	t.Run("create", func(t *testing.T) {
		s := NewStore()
		p, err := s.CreateOrGet(&Payment{
			Id:                  "1234",
			state:               PaymentStateNew,
			Version:             "c415e106-4183-4c40-94cd-383eeb9a7704",
			Amount:              1050,
			Currency:            "GBP",
			CardNumber:          "1234123412341234",
			ExpiryDate:          "1220",
			CardHolder:          "John Doe",
			Cvv:                 "123",
			Expected3dsResponse: "foo",
		})
		assert.NoError(t, err)
		assert.NotNil(t, p)

		created := *p

		p, err = s.Get("1234")
		assert.NoError(t, err)
		assert.NotNil(t, p)
		retrieved := *p

		for _, p := range []*Payment{&created, &retrieved} {
			assert.Equal(t, PaymentId("1234"), p.Id)
			assert.Equal(t, PaymentStateNew, p.state)
			assert.Equal(t, "c415e106-4183-4c40-94cd-383eeb9a7704", p.Version)
			assert.Equal(t, int64(1050), p.Amount)
			assert.Equal(t, "GBP", p.Currency)
			assert.Equal(t, "1234123412341234", p.CardNumber)
			assert.Equal(t, "1220", p.ExpiryDate)
			assert.Equal(t, "John Doe", p.CardHolder)
			assert.Equal(t, "123", p.Cvv)
			assert.Equal(t, "foo", p.Expected3dsResponse)
		}
	})

	t.Run("create same id", func(t *testing.T) {
		s := NewStore()
		p, err := s.CreateOrGet(&Payment{
			Id:      "1234",
			state:   PaymentStateNew,
			Version: "c415e106-4183-4c40-94cd-383eeb9a7704",
		})
		assert.NoError(t, err)
		assert.NotNil(t, p)

		p, err = s.CreateOrGet(&Payment{
			Id:      "1234",
			state:   PaymentState3dSecureRequired,
			Version: "81628aa6-9841-4d78-a0c1-6c793ac15d58",
		})
		assert.NoError(t, err)
		assert.NotNil(t, p)

		// Original payment data should be returned.
		assert.Equal(t, PaymentId("1234"), p.Id)
		assert.Equal(t, PaymentStateNew, p.state)
		assert.Equal(t, "c415e106-4183-4c40-94cd-383eeb9a7704", p.Version)
	})
}

func Test_paymentStoreImpl_Get(t *testing.T) {
	t.Run("get existing", func(t *testing.T) {
		s := NewStore()
		_, err := s.CreateOrGet(&Payment{
			Id:      "1234",
			state:   PaymentStateConfirmed,
			Version: "c415e106-4183-4c40-94cd-383eeb9a7704",
			Amount:  1050,
		})
		assert.NoError(t, err)

		p, err := s.Get("1234")
		assert.NoError(t, err)
		assert.Equal(t, PaymentId("1234"), p.Id)
		assert.Equal(t, "c415e106-4183-4c40-94cd-383eeb9a7704", p.Version)
	})

	t.Run("get not found", func(t *testing.T) {
		s := NewStore()
		_, err := s.Get("1111")
		assert.Error(t, err)
	})
}

func Test_paymentStoreImpl_Update(t *testing.T) {
	t.Run("update existing", func(t *testing.T) {
		s := NewStore()
		_, err := s.CreateOrGet(&Payment{
			Id:      "1234",
			state:   PaymentStateConfirmed,
			Version: "c415e106-4183-4c40-94cd-383eeb9a7704",
			Amount:  1050,
		})
		assert.NoError(t, err)

		p1, err := s.Update("1234", "c415e106-4183-4c40-94cd-383eeb9a7704", func(p *Payment) error {
			p.Amount = 999
			return nil
		})
		assert.NoError(t, err)

		// version should be updated
		assert.NotEqual(t, "c415e106-4183-4c40-94cd-383eeb9a7704", p1.Version)
		assert.Equal(t, int64(999), p1.Amount)

		p2, err := s.Get("1234")
		assert.NoError(t, err)

		assert.Equal(t, p1.Version, p2.Version)
		assert.Equal(t, p1.Amount, p2.Amount)
	})

	t.Run("update missing", func(t *testing.T) {
		s := NewStore()
		_, err := s.CreateOrGet(&Payment{Id: "1234", Version: "c415e106-4183-4c40-94cd-383eeb9a7704"})
		assert.NoError(t, err)

		_, err = s.Update("12345", "2cea903d-b7f4-4f2c-a39e-0b4a71ff5b2a", func(p *Payment) error {
			p.Amount = 999
			return nil
		})
		assert.ErrorContains(t, err, "not found")
	})

	t.Run("update version mismatch", func(t *testing.T) {
		s := NewStore()
		_, err := s.CreateOrGet(&Payment{
			Id:      "1234",
			state:   PaymentStateConfirmed,
			Version: "c415e106-4183-4c40-94cd-383eeb9a7704",
			Amount:  1050,
		})
		assert.NoError(t, err)

		_, err = s.Update("1234", "2cea903d-b7f4-4f2c-a39e-0b4a71ff5b2a", func(p *Payment) error {
			p.Amount = 999
			return nil
		})
		assert.ErrorContains(t, err, "version mismatch")
	})
}

func Test_paymentStoreImpl_List(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		s := NewStore()

		for _, state := range []PaymentState{PaymentStateConfirmed, PaymentStateNew, PaymentState3dSecureRequired, PaymentStateConfirmed} {
			_, err := s.CreateOrGet(&Payment{
				Id:      PaymentId(uuid.NewString()),
				state:   state,
				Version: string(state),
			})
			assert.NoError(t, err)
		}

		ps, err := s.List(PaymentStateConfirmed)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(ps))
	})
}
