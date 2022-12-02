package resources

import (
	"mkuznets.com/go/upsp/gateway/models"
	"strings"
)

func PaymentModelToResource(p *models.Payment) *PaymentResource {
	return &PaymentResource{
		Id:    p.Id,
		State: p.State,

		Amount:   p.Amount,
		Currency: p.Currency,

		CardNumber: strings.Repeat("*", len(p.CardNumber)-4) + p.CardNumber[len(p.CardNumber)-4:],
		ExpiryDate: p.ExpiryDate,
		CardHolder: p.CardHolder,
		Cvv:        strings.Repeat("*", len(p.Cvv)),

		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
