package rest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/greenboxal/agibootstrap/pkg/psi/psiml"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
	"github.com/greenboxal/agibootstrap/psidb/services/indexing"
	"github.com/greenboxal/agibootstrap/psidb/services/search"
)

type RenderRequest struct {
	Text string `json:"text"`

	ReturnEmbeddings bool `json:"return_embeddings"`
}

type RenderResponse struct {
	Rendered   string      `json:"rendered"`
	Embeddings [][]float32 `json:"embeddings,omitempty"`
}

type RenderHandler struct {
	core     coreapi.Core
	search   *search.SearchService
	embedder indexing.NodeEmbedder
}

func NewRenderHandler(
	core coreapi.Core,
	search *search.SearchService,
	embedder indexing.NodeEmbedder,
) *RenderHandler {
	return &RenderHandler{
		core:     core,
		search:   search,
		embedder: embedder,
	}
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

	if req.ReturnEmbeddings {
		embeddings, err := r.embedder.EmbeddingsForNode(request.Context(), stdlib.NewText(res.Rendered))

		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if embeddings.Next() {
			res.Embeddings = append(res.Embeddings, embeddings.Value().ToFloat32Slice(nil))
		}
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(writer).Encode(res); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}
