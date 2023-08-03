package rpc

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/swaggest/jsonrpc"
	"github.com/swaggest/swgui"
	"github.com/swaggest/swgui/v3cdn"
)

type Docs struct {
	chi.Router
}

func NewDocs(rpc *RpcService) *Docs {
	mux := &Docs{Router: chi.NewMux()}

	mux.Method(http.MethodGet, "/openapi.json", rpc.OpenAPI)

	mux.Mount("/", v3cdn.NewHandlerWithConfig(swgui.Config{
		Title:       "RPC",
		SwaggerJSON: "/rpc/v1/docs/openapi.json",
		BasePath:    "/rpc/v1/docs",
		SettingsUI:  jsonrpc.SwguiSettings(nil, "/rpc/v1"),
	}))

	return mux
}
