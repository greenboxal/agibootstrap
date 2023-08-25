package mgmtv1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/greenboxal/agibootstrap/psidb/services/pubsub"
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

	r.Get("/scheduler", r.handleScheduler)

	return r
}

func (r *Router) handleScheduler(writer http.ResponseWriter, request *http.Request) {
	stats := r.ps.Scheduler().DumpStatistics()
	data, err := json.Marshal(stats)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/json")
	_, _ = writer.Write(data)
}
