package rest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/greenboxal/agibootstrap/pkg/psi/psiml"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/services"
)

type RenderRequest struct {
	Text string `json:"text"`
}

type RenderResponse struct {
	Rendered string `json:"rendered"`
}

type RenderHandler struct {
	core   coreapi.Core
	search *services.SearchService
}

func NewRenderHandler(core coreapi.Core, search *services.SearchService) *RenderHandler {
	return &RenderHandler{core: core, search: search}
}

func (r *RenderHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var req RenderRequest
	var res RenderResponse

	data, err := io.ReadAll(request.Body)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(data, &req); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	err = r.core.RunTransaction(request.Context(), func(ctx context.Context, tx coreapi.Transaction) error {
		processor := psiml.NewTextProcessor(tx.Graph(), r.search)
		rendered, err := processor.Process(ctx, req.Text)

		if err != nil {
			return err
		}

		res.Rendered = rendered

		return nil
	})

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(writer).Encode(res); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}
