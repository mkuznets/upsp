package gateway

import (
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"mkuznets.com/go/gateway/gateway/models"
	"time"
)

type CreatePaymentRequest struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`

	CardNumber string `json:"card_number"`
	ExpiryDate string `json:"expiry_date"`
	CardHolder string `json:"card_holder"`
	Cvv        string `json:"cvv"`
}

func isExpiryDate(value interface{}) error {
	t, err := time.Parse("0106", value.(string))
	if err != nil {
		return fmt.Errorf("invalid expiry date")
	}
	if t.Before(time.Now()) {
		return fmt.Errorf("expiry date is in the past")
	}
	return nil
}

func (r *CreatePaymentRequest) Validate() error {
	return validation.ValidateStruct(
		r,
		validation.Field(&r.Amount, validation.Required, validation.Min(1), validation.Max(99999999)),
		validation.Field(&r.Currency, validation.Required, is.CurrencyCode),
		validation.Field(&r.CardNumber, validation.Required, validation.Length(16, 16), is.CreditCard),
		validation.Field(&r.ExpiryDate, validation.Required, validation.Length(4, 4), validation.By(isExpiryDate)),
		validation.Field(&r.CardHolder, validation.Required, validation.Length(1, 999)),
		validation.Field(&r.Cvv, validation.Required, validation.Length(3, 4)),
	)
}

type CreatePaymentResponse struct {
	Id    string              `json:"id"`
	State models.PaymentState `json:"state"`
}

type PaymentResource struct {
	Id    string              `json:"id"`
	State models.PaymentState `json:"state"`

	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`

	CardNumber string `json:"number"`
	ExpiryDate string `json:"expiry_date"`
	CardHolder string `json:"holder"`
	Cvv        string `json:"cvv"`
}
