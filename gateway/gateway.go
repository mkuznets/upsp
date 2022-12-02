package gateway

import (
	"context"
	"mkuznets.com/go/upsp/acquirer"
	"mkuznets.com/go/upsp/gateway/api"
	"mkuznets.com/go/upsp/gateway/store"
	"mkuznets.com/go/upsp/gateway/transitioner"
)

type Gateway interface {
	// Start starts the gateway.
	Start(ctx context.Context)
}

type gatewayImpl struct {
	api          *api.Api
	store        store.Store
	transitioner transitioner.Transitioner
}

func New(store store.Store, acq acquirer.Acquirer) Gateway {
	return &gatewayImpl{
		store:        store,
		api:          api.New(store, acq),
		transitioner: transitioner.New(store, acq),
	}
}

func (g *gatewayImpl) Start(ctx context.Context) {
	go g.transitioner.Start(ctx)
	g.api.Start(ctx)
}
