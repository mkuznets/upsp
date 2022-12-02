package acquirer

type PaymentResource struct {
	Id      PaymentId
	State   PaymentState
	Version string
}

type CreatePaymentRequest struct {
	Id       PaymentId
	Amount   int64
	Currency string
}

type CreatePaymentResponse struct {
	PaymentResource
}

type AuthorisePaymentRequest struct {
	CardNumber string
	ExpiryDate string
	CardHolder string
	Cvv        string
}

type AuthorisePaymentResponse struct {
	Payment PaymentResource
	AuthUrl string
}

type Submit3dSecureRequest struct {
	Token string
}

type Submit3dSecureResponse struct {
	Payment PaymentResource
}

type ConfirmPaymentResponse struct {
	Payment PaymentResource
}

type CancelPaymentResponse struct {
	Payment PaymentResource
}
