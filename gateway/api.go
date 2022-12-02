package gateway

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"log"
	"mkuznets.com/go/gateway/gateway/models"
	"mkuznets.com/go/gateway/gateway/store"
	"net/http"
	"time"
)

type Api struct {
	addr   string
	store  store.Store
	router *chi.Mux
}

func NewApi(store store.Store) *Api {
	a := &Api{
		addr:   ":8080",
		store:  store,
		router: chi.NewRouter(),
	}

	a.router.Use(middleware.Timeout(30 * time.Second))
	a.router.Use(middleware.Recoverer)

	a.router.Route("/payments", func(r chi.Router) {
		r.Post("/", a.CreatePayment)
		r.Get("/{paymentId}", a.GetPayment)
	})

	return a
}

func (api *Api) Start() {
	if err := http.ListenAndServe(api.addr, api.router); err != nil {
		log.Printf("[WARN] server has terminated: %s", err)
	}
}

func (api *Api) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var request CreatePaymentRequest
	if err := render.DecodeJSON(r.Body, &request); err != nil {
		Handle(w, r, New(err, http.StatusBadRequest, "Invalid request"))
		return
	}

	if err := request.Validate(); err != nil {
		Handle(w, r, New(err, http.StatusBadRequest, err.Error()))
		return
	}

	paymentModel := &models.Payment{
		Id:         uuid.NewString(),
		MerchantId: "123",
		Amount:     request.Amount,
		Currency:   request.Currency,
		State:      models.PaymentStateNew,
		Version:    uuid.NewString(),
		CardNumber: request.CardNumber,
		CardHolder: request.CardHolder,
		ExpiryDate: request.ExpiryDate,
		Cvv:        request.Cvv,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	ctx := r.Context()

	err := api.store.Db().Tx(ctx, func(tx store.Tx) error {
		id, err := api.store.Payments().Create(ctx, paymentModel)
		if err != nil {
			return err
		}
		p, err := api.store.Payments().Get(ctx, id)
		if err != nil {
			return err
		}

		resp := &CreatePaymentResponse{
			Id:    p.Id,
			State: p.State,
		}
		render.Status(r, http.StatusCreated)
		render.JSON(w, r, resp)
		return nil
	})
	if err != nil {
		Handle(w, r, err)
		return
	}

	return
}

func (api *Api) GetPayment(w http.ResponseWriter, r *http.Request) {
	paymentId := chi.URLParam(r, "paymentId")
	p, err := api.store.Payments().Get(r.Context(), paymentId)
	if err != nil {
		Handle(w, r, err)
		return
	}

	resp := &PaymentResource{
		Id:         p.Id,
		State:      p.State,
		Amount:     p.Amount,
		Currency:   p.Currency,
		CardNumber: p.CardNumber,
		ExpiryDate: p.ExpiryDate,
		CardHolder: p.CardHolder,
		Cvv:        p.Cvv,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
	return
}

type APIError struct {
	Err  error
	Code int
	Msg  string
}

func (e *APIError) Error() string {
	return e.Msg
}

func (e *APIError) JSON() render.M {
	return render.M{
		"error":   http.StatusText(e.Code),
		"message": e.Msg,
	}
}

func New(err error, code int, msg string) *APIError {
	return &APIError{err, code, msg}
}

func Handle(w http.ResponseWriter, r *http.Request, err error) {
	switch v := err.(type) {
	case *APIError:
		render.Status(r, v.Code)
		render.JSON(w, r, v.JSON())
	default:
		//if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
		//	hub.CaptureException(err)
		//}
		log.Printf("[ERR] %v", err)
		e := New(err, http.StatusInternalServerError, "Unexpected system error")
		render.Status(r, e.Code)
		render.JSON(w, r, e.JSON())
	}
}
