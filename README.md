# µPSP

µPSP is a toy payment service provider written in Golang. It consists of the following components:

* **acquirer** is an embedded in-memory acquiring bank simulator. It facilitates card payment initiation and tracking,
  and partially implements its lifecycle, including 3DS authorisation and refunds. The acquirer also provides test card
  numbers that simulate certain payment outcomes.
* **gateway** provides a simple REST API to initiate and track a card payment using the acquirer.

## Acquirer

The acquirer implements the following payment flow:

![Acquiring payment flow](assets/acquirer.png)

### API

The acquirer provides the following interface:

```
// Start runs background asynchronous tasks, in particular test refund scenarios and 3DS timeouts.
Start()

// GetPayment returns the current payment details and status.
GetPayment(PaymentId) (*PaymentResource, error)

// CreatePayment creates a new payment for the given amount and currency.
CreatePayment(*CreatePaymentRequest) (*CreatePaymentResponse, error)

// AuthorisePayment stores the provided payment method details and initiates the payment autorisation.
AuthorisePayment(id PaymentId, version string, req *AuthorisePaymentRequest) (*AuthorisePaymentResponse, error)

// Submit3dSecure 
Submit3dSecure(id PaymentId, version string, req *Submit3dSecureRequest) (*Submit3dSecureResponse, error)

// ConfirmPayment finalises the charge of the given authorised payment.
ConfirmPayment(id PaymentId, version string) (*ConfirmPaymentResponse, error)

// CancelPayment cancels the given payment. The resulting payment state varies depending on the current state.
// Initiates a payment refund for confirmed payments.
CancelPayment(id PaymentId, version string) (*CancelPaymentResponse, error)
```

Each payment mutation has a `version` argument. The version is a UUIDv4 that is stored on the payment and regenerated on
each payment mutation. If the version provided in the operation argument does not match the current
one, the operation fails. This forces the consumer to reload the payments to avoid stale updates.

### Test Cards

The acquirer includes a number of hard-coded card numbers that trigger testing scenarios for both happy and unhappy path
of a card payment:

* 3DS required, successful authorisation: 4000000000003220, 4000000000003063
* 3DS required, failed authorisation: 4000000000003097, 4000008400001280
* No 3DS required:
    * Successful authorisation: 4242424242424242, 5555555555554444
    * Successful authorisation, auto-refund after a certain timeout: 4000000000005126, 4000000000007726

### Implementation Details

* Payments are stored in memory using a lock-protected map.
* For simplicity, only pull-based payment tracking is implemented. Ideally, the acquirer should also be able to send
  asynchronous events or webhooks on each payment update.
* The current version does not implement a mock service to trigger 3DS verification.

## Gateway

###                                     