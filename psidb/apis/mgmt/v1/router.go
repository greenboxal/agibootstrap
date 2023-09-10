package mgmtv1

import (
	"github.com/go-chi/chi/v5"

	"github.com/greenboxal/agibootstrap/psidb/core/pubsub"
)

type Router struct {
	chi.Router

	ps *pubsub.Manager
}

func NewRouter(ps *pubsub.Manager) *Router {
	r := &Router{
		Router: chi.NewRouter(),

		ps: ps,
	}

	return r
}
