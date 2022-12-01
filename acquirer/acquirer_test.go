package acquirer

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"mkuznets.com/go/gateway/acquirer/models"
)

func TestNewAcquirer(t *testing.T) {
	acq, err := NewAcquirer()
	assert.NoError(t, err)

	r, err := acq.CreatePayment(&CreatePaymentRequest{
		Id:       models.PaymentId(uuid.NewString()),
		Amount:   100,
		Currency: "GBP",
		HookUrl:  "http://127.0.0.1",
	})
	assert.NoError(t, err)
	fmt.Println(r)

	rr, err := acq.AuthorisePayment(r.Id, r.Version, &AuthorisePaymentRequest{
		CardNumber: "4111111111111111",
		ExpiryDate: "1122",
		CardHolder: "John Doe",
		Cvv:        "233",
	})
	assert.NoError(t, err)
	fmt.Println(rr)

	rrr, err := acq.Submit3dSecure(r.Id, rr.Payment.Version, &Submit3dSecureRequest{
		Token: "123456",
	})
	assert.NoError(t, err)
	fmt.Println(rrr)
}
