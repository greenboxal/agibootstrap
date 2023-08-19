package openaiv1

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	openai2 "github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"

	gpt2 "github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/psidb/modules/gpt"
)

type Router struct {
	chi.Router

	client *openai2.Client
	ecm    *gpt.EmbeddingCacheManager

	rp *httputil.ReverseProxy
}

func NewRouter(ecm *gpt.EmbeddingCacheManager) *Router {
	router := &Router{
		Router: chi.NewRouter(),
		client: gpt2.GlobalClient,
		ecm:    ecm,
	}

	router.rp = &httputil.ReverseProxy{
		Rewrite: func(request *httputil.ProxyRequest) {
			if request.Out.Header.Get("Authorization") == "" {
				request.Out.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))
			}

			request.SetURL(&url.URL{
				Scheme: "https",
				Host:   "api.openai.com",
				Path:   "/v1",
			})
		},
	}

	router.Get("/models", buildHandler(router.ListModels))
	router.Get("/models/:id", buildHandler(router.GetModel))
	router.Post("/embeddings", buildHandler(router.CreateEmbeddings))
	router.Post("/completions", buildHandler(router.CreateCompletion))
	router.Post("/chat/completions", buildHandler(router.CreateChatCompletion))

	router.NotFound(func(writer http.ResponseWriter, request *http.Request) {
		router.rp.ServeHTTP(writer, request)
	})

	return router
}

func (r *Router) ListModels(req *Request[struct{}], writer *ResponseWriter) error {
	models, err := r.client.ListModels(req.Context())

	if err != nil {
		return err
	}

	writer.WriteResponse(models)

	return nil
}

func (r *Router) GetModel(req *Request[struct{}], writer *ResponseWriter) error {
	model, err := r.client.GetModel(req.Context(), chi.URLParam(req.Request, "id"))

	if err != nil {
		return err
	}

	writer.WriteResponse(model)

	return nil
}

func (r *Router) CreateEmbeddings(req *Request[openai.EmbeddingRequest], writer *ResponseWriter) error {
	embedder := r.ecm.GetEmbedder(&openai2.Embedder{Client: r.client, Model: req.Payload.Model})
	embedding, err := embedder.GetEmbeddings(req.Context(), req.Payload.Input)

	if err != nil {
		return err
	}

	data := lo.Map(embedding, func(e llm.Embedding, i int) openai.Embedding {
		return openai.Embedding{
			Object:    "embedding",
			Embedding: e.Embeddings,
			Index:     i,
		}
	})

	writer.WriteResponse(openai.EmbeddingResponse{
		Object: "list",
		Model:  req.Payload.Model,
		Data:   data,
	})

	return nil
}

func (r *Router) CreateCompletion(req *Request[openai.CompletionRequest], writer *ResponseWriter) error {
	if req.Payload.Stream {
		result, err := r.client.CreateCompletionStream(req.Context(), req.Payload)

		if err != nil {
			return err
		}

		defer result.Close()

		for {
			chunk, err := result.Recv()

			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}

			data, err := json.Marshal(chunk)

			if err != nil {
				return err
			}

			if _, err := writer.Write(data); err != nil {
				return err
			}
		}
	} else {
		result, err := r.client.CreateCompletion(req.Context(), req.Payload)

		if err != nil {
			return err
		}

		writer.WriteResponse(result)
	}

	return nil
}

func (r *Router) CreateChatCompletion(req *Request[openai.ChatCompletionRequest], writer *ResponseWriter) error {
	if req.Payload.Stream {
		result, err := r.client.CreateChatCompletionStream(req.Context(), req.Payload)

		if err != nil {
			return err
		}

		defer result.Close()

		for {
			chunk, err := result.Recv()

			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}

			data, err := json.Marshal(chunk)

			if err != nil {
				return err
			}

			if _, err := writer.Write(data); err != nil {
				return err
			}
		}
	} else {
		result, err := r.client.CreateChatCompletion(req.Context(), req.Payload)

		if err != nil {
			return err
		}

		writer.WriteResponse(result)
	}

	return nil
}
