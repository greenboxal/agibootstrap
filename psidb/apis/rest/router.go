package rest

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Router struct {
	chi.Router
}

func NewRouter(
	resourceHandler *ResourceHandler,
	searchHandler *SearchHandler,
) *Router {
	router := &Router{
		Router: chi.NewRouter(),
	}

	router.Mount("/psi", http.StripPrefix("/psi", resourceHandler))
	router.Mount("/search", searchHandler)

	return router
}
