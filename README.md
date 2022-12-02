# µPSP

µPSP is a toy payment service provider written in Golang. It consists of the following components:

* **acquirer** is an embedded in-memory acquiring bank simulator. It facilitates card payment initiation and tracking,
  and partially implements its lifecycle, including 3DS authorisation and refunds. The acquirer also provides test card
  numbers that simulate certain payment outcomes.
* **gateway** provides a simple REST API to initiate and track a card payment using the acquirer. Even if the acquirer
  fails in the middle of a transition sequence, the gateway will pick up the pending transaction and update its state
  asynchronously.

## Usage

```bash
$ # Build and run directly (requires a Postgres instance with the schema from ./sql)
$ go build mkuznets.com/go/upsp/cmd/upsp
$ ./upsp -p 'postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable'

$ # Build and run using docker-compose (automatically initialises Postgres):
$ docker-compose build
$ docker-compose up
```

In both cases the gateway API will be available on 127.0.0.1:8080

```bash
$ curl -sX "POST" "http://127.0.0.1:8080/payments" \
     -H 'Content-Type: application/json; charset=utf-8' \
     -d $'{
  "amount": 10,
  "currency": "EUR",
  "card_number": "4000008400001280",
  "card_holder": "Jane Doe",
  "cvv": "123",
  "expiry_date": "0123"
}' | jq
{
  "id": "3eca1f99-f927-4c36-9c3f-8c5ff9cba6f5",
  "state": "action_required",
  "amount": 10,
  "currency": "EUR",
  "card_number": "************1280",
  "card_holder": "Jane Doe",
  "expiry_date": "0123",
  "cvv": "***",
  "created_at": "2022-12-02T10:15:47.187559Z",
  "updated_at": "2022-12-02T10:15:47.192876Z"
}

$ curl -sX "POST" "http://127.0.0.1:8080/payments" \
     -H 'Content-Type: application/json; charset=utf-8' \
     -d $'{
  "amount": -10,
  "currency": "EEE",
  "card_number": "4000008400001111",
  "card_holder": "Jane Doe",
  "cvv": "",
  "expiry_date": "0122"
}' | jq
{
  "error": "Bad Request",
  "message": "amount: must be no less than 1; card_number: must be a valid credit card number; currency: must be valid ISO 4217 currency code; cvv: cannot be blank; expiry_date: expiry date is in the past."
}

$ curl -s "http://127.0.0.1:8080/payments/e048708e-1e51-429d-80af-dd01c8353cfd" | jq
{
  "id": "e048708e-1e51-429d-80af-dd01c8353cfd",
  "state": "rejected",
  "amount": 10,
  "currency": "EUR",
  "card_number": "************1280",
  "card_holder": "Jane Doe",
  "expiry_date": "0123",
  "cvv": "***",
  "created_at": "2022-12-02T08:32:17.530203Z",
  "updated_at": "2022-12-02T08:33:23.327842Z"
}
```

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

All mutation operations are idempotent:

* `CreatePayment` is deduplicated based on the payment ID.
* Updated and transitions are deduplicated via the `version` argument. The version is a UUIDv4 that is stored on the
  payment and regenerated on
  each payment mutation. If the version provided in the operation argument does not match the current
  one, the operation fails.

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

The gateway provides a REST API to initiate payments, execute them in the acquirer, and track their status. The payment
lifecycle exposed in the REST API is slightly simpler:

![Gateway payment flow](assets/gateway.png)

Note that while these transitions are coded are valid, not all of them can be initiated in the current implementation.

### API

#### Payment Initiation

`POST /payments`

Request:

```
{
  "amount": 10,                           // Payment amount in minor units (e.g. cents)
  "currency": "EUR",                      // ISO currency code
  // Payment method details:
  "card_number": "4000008400001280",
  "card_holder": "Jane Doe",
  "cvv": "123",
  "expiry_date": "0123"
}
```

Response:

```
{
  "id": "<payment UUID>",
  "state": "<processing|action_required|paid|rejected|refunded|cancelled>",
  "amount": 10,
  "currency": "EUR",
  "card_number": "************9999",
  "card_holder": "Jane Doe",
  "expiry_date": "0123",
  "cvv": "***",
  "created_at": "<ISO time>",
  "updated_at": "<ISO time>"
}
```

#### Payment Tracking

`GET /payments/<payment UUID>`

Response:

```
{
  "id": "<payment UUID>",
  "state": "<processing|action_required|paid|rejected|refunded|cancelled>",
  "amount": 10,
  "currency": "EUR",
  "card_number": "************9999",
  "card_holder": "Jane Doe",
  "expiry_date": "0123",
  "cvv": "***",
  "created_at": "<ISO time>",
  "updated_at": "<ISO time>"
}
```

### Implementation Details

* The Gateway uses PostgreSQL to store payments. Operations with payments are abstracted in a repository-like
  interface `store.Payments` with the common database abstraction `store.Store` that implements PG transactions so that
  they can potentially span across multiple repositories (see `Store.Tx()`).
* The `Transitioner` interface implements a synchronous payment transition. When the payment reaches a terminal state
  or "action_required", the `Transition` method reads the upstream state from the acquirer and updates the record
  accordingly. It also implements a background worker that regularly tries to transition all gateway payments.
