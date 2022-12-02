package acquirer

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewAcquirer(t *testing.T) {
	acq := New(NewStore())
	_, err := acq.CreatePayment(&CreatePaymentRequest{
		Id:       PaymentId(uuid.NewString()),
		Amount:   100,
		Currency: "GBP",
	})
	assert.NoError(t, err)
}

func TestAcquirer_CreatePayment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		storeMock := &StoreMock{
			CreateOrGetFunc: func(payment *Payment) (*Payment, error) {
				assert.Equal(t, PaymentId("f6fee5b0-c126-4889-aeb3-b9fb1c8b3a04"), payment.Id)
				assert.Equal(t, int64(100), payment.Amount)
				assert.Equal(t, "GBP", payment.Currency)
				assert.Equal(t, PaymentStateNew, payment.State())
				return payment, nil
			},
		}

		acq := New(storeMock)
		py, err := acq.CreatePayment(&CreatePaymentRequest{
			Id:       "f6fee5b0-c126-4889-aeb3-b9fb1c8b3a04",
			Amount:   100,
			Currency: "GBP",
		})
		assert.NoError(t, err)
		assert.NotNil(t, py)
		assert.Len(t, storeMock.calls.CreateOrGet, 1)

		assert.Equal(t, PaymentId("f6fee5b0-c126-4889-aeb3-b9fb1c8b3a04"), py.Id)
		assert.Equal(t, PaymentStateNew, py.State)
		assert.NotEmpty(t, py.Version)
	})
}

func TestAcquirer_AuthorisePayment(t *testing.T) {
	t.Run("mock", func(t *testing.T) {
		storeMock := &StoreMock{
			UpdateFunc: func(id PaymentId, version string, fn func(*Payment) error) (*Payment, error) {
				var p Payment
				err := fn(&p)
				assert.ErrorContains(t, err, "is not in new")

				_ = p.SetState(PaymentStateNew)
				err = fn(&p)
				assert.NoError(t, err)

				assert.Equal(t, "4242424242424242", p.CardNumber)
				assert.Equal(t, "1077", p.ExpiryDate)
				assert.Equal(t, "John Doe", p.CardHolder)
				assert.Equal(t, "123", p.Cvv)

				return &p, nil
			},
		}

		acq := New(storeMock)

		_, _ = acq.AuthorisePayment("123", "f6fee5b0-c126-4889-aeb3-b9fb1c8b3a04", &AuthorisePaymentRequest{
			CardNumber: "4242424242424242",
			ExpiryDate: "1077",
			CardHolder: "John Doe",
			Cvv:        "123",
		})
		assert.Len(t, storeMock.calls.Update, 1)
	})

	t.Run("auth results", func(t *testing.T) {
		states := make([]PaymentState, 0)

		for _, card := range []string{"4242424242424242", "4000008400001280", "4000000000000101", "4000000000000000"} {
			acq := New(NewStore())
			py, err := acq.CreatePayment(&CreatePaymentRequest{
				Id:       "f6fee5b0-c126-4889-aeb3-b9fb1c8b3a04",
				Amount:   100,
				Currency: "GBP",
			})
			require.NoError(t, err)

			resp, err := acq.AuthorisePayment(py.Id, py.Version, &AuthorisePaymentRequest{
				CardNumber: card,
				ExpiryDate: "1077",
				CardHolder: "John Doe",
				Cvv:        "123",
			})
			require.NoError(t, err)
			states = append(states, resp.Payment.State)
		}

		assert.Equal(t, []PaymentState{"authorised", "3d_secure_required", "rejected", "rejected"}, states)
	})
}
