package api

import (
	"context"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"log"
	"mkuznets.com/go/upsp/acquirer"
	"mkuznets.com/go/upsp/gateway/models"
	"mkuznets.com/go/upsp/gateway/store"
	"mkuznets.com/go/upsp/gateway/transitioner"
	"net/http"
	"time"
)

type Api struct {
	addr         string
	store        store.Store
	transitioner transitioner.Transitioner
	router       *chi.Mux
}

func New(store store.Store, acq acquirer.Acquirer) *Api {
	a := &Api{
		addr:         ":8080",
		store:        store,
		router:       chi.NewRouter(),
		transitioner: transitioner.New(store, acq),
	}

	a.router.Use(middleware.Timeout(30 * time.Second))
	a.router.Use(middleware.Recoverer)
	a.router.Use(middleware.Logger)

	a.router.Route("/payments", func(r chi.Router) {
		r.Post("/", a.CreatePayment)
		r.Get("/{paymentId}", a.GetPayment)
	})

	return a
}

func (api *Api) Start(ctx context.Context) {
	if err := http.ListenAndServe(api.addr, api.router); err != nil {
		log.Printf("[WARN] server has terminated: %s", err)
	}
}

func (api *Api) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var request CreatePaymentRequest
	if err := render.DecodeJSON(r.Body, &request); err != nil {
		renderApiError(w, r, err, http.StatusBadRequest, "invalid request")
		return
	}

	if err := request.Validate(); err != nil {
		renderApiError(w, r, err, http.StatusBadRequest, err.Error())

		renderError(w, r, &Error{err, http.StatusBadRequest, err.Error()})
		return
	}

	paymentModel := &models.Payment{
		Id:         uuid.NewString(),
		Amount:     request.Amount,
		Currency:   request.Currency,
		State:      models.PaymentStateProcessing,
		CardNumber: request.CardNumber,
		CardHolder: request.CardHolder,
		ExpiryDate: request.ExpiryDate,
		Cvv:        request.Cvv,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	ctx := r.Context()

	err := api.store.Tx(ctx, func(ctx context.Context) error {
		id, err := api.store.Payments().Create(ctx, paymentModel)
		if err != nil {
			return err
		}

		if err := api.transitioner.Transition(ctx, id); err != nil {
			log.Printf("[ERR] %v", err)
		}

		p, err := api.store.Payments().Get(ctx, id)
		if err != nil {
			return err
		}

		resp := &CreatePaymentResponse{
			PaymentResource: *PaymentModelToResource(p),
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, resp)
		return nil
	})
	if err != nil {
		renderError(w, r, err)
		return
	}

	return
}

func (api *Api) GetPayment(w http.ResponseWriter, r *http.Request) {
	paymentId := chi.URLParam(r, "paymentId")
	p, err := api.store.Payments().Get(r.Context(), paymentId)
	switch {
	case err == pgx.ErrNoRows:
		e := fmt.Errorf("no payment found")
		renderApiError(w, r, e, http.StatusNotFound, e.Error())
		return
	case err != nil:
		renderError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, PaymentModelToResource(p))
	return
}
