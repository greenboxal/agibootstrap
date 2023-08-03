package gateway

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	cid2 "github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/graphindex"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/graphstore"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/apis/rest"
	"github.com/greenboxal/agibootstrap/psidb/core"
)

var logger = logging.GetLogger("apis/gateway")

type Gateway struct {
	chi.Router

	rootPath     psi.Path

	indexManager *graphindex.Manager

	core         *core.Core
}

func NewGateway(
	core *core.Core,
	indexManager *graphindex.Manager,
	root psi.Path,
) *Gateway {
	gw := &Gateway{
		core: core,
		indexManager: indexManager,
		rootPath:     root,
		router:       chi.NewRouter(),
	}

	gw.router.Get("/_objects/{cid}", gw.handleObjectStoreGet)
	gw.router.HandleFunc("/_search", gw.handleSearch)
	gw.router.Mount("/v1", http.StripPrefix("/v1", rest.NewRouter(graph.)))

	gw.router.Route("/psi", func(r chi.Router) {
		r.NotFound(func(writer http.ResponseWriter, request *http.Request) {
			request.URL.Path = strings.TrimPrefix(request.URL.Path, "/psi")
			request.URL.Path = strings.TrimPrefix(request.URL.Path, "/")

			gw.handlePsiDb(writer, request)
		})
	})

	return gw
}


func (gw *Gateway) handleObjectStoreGet(writer http.ResponseWriter, request *http.Request) {
	cidStr := chi.URLParam(request, "cid")
	cid, err := cid2.Parse(cidStr)

	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(writer, err.Error())
		return
	}

	link := cidlink.Link{Cid: cid}
	obj, err := gw.graph.LinkSystem().LoadRaw(linking.LinkContext{Ctx: request.Context()}, link)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(writer, err.Error())
		return
	}

	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(obj)
}
