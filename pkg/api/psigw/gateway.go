package psigw

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
)

var logger = logging.GetLogger("api/psigw")

type Gateway struct {
	router       chi.Router
	server       http.Server
	project      project.Project
	graph        *graphstore.IndexedGraph
	indexManager *graphindex.Manager
	rootPath     psi.Path
}

func NewGateway(
	prj project.Project,
	graph *graphstore.IndexedGraph,
	indexManager *graphindex.Manager,
	root psi.Path,
) *Gateway {
	gw := &Gateway{
		project:      prj,
		graph:        graph,
		indexManager: indexManager,
		rootPath:     root,
		router:       chi.NewRouter(),
	}

	gw.server.Handler = gw.router

	gw.router.Use(middleware.RealIP)
	gw.router.Use(middleware.RequestID)
	gw.router.Use(middleware.Logger)
	gw.router.Use(middleware.Recoverer)

	gw.router.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	gw.router.Get("/_objects/{cid}", gw.handleObjectStoreGet)
	gw.router.HandleFunc("/_search", gw.handleSearch)

	gw.router.Route("/psi", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			})
		})

		r.NotFound(func(writer http.ResponseWriter, request *http.Request) {
			request.URL.Path = strings.TrimPrefix(request.URL.Path, "/psi")
			request.URL.Path = strings.TrimPrefix(request.URL.Path, "/")

			gw.handlePsiDb(writer, request)
		})
	})

	return gw
}

func (gw *Gateway) Start(ctx context.Context) error {
	endpoint := os.Getenv("AGIB_LISTEN_ENDPOINT")

	if endpoint == "" {
		endpoint = "0.0.0.0:22440"
	}

	l, err := net.Listen("tcp", endpoint)

	if err != nil {
		return err
	}

	logger.Infow("Server is listening", "endpoint", endpoint)

	go func() {
		if err := gw.server.Serve(l); err != nil {
			logger.Error(err)
		}
	}()

	return nil
}

func (gw *Gateway) Shutdown(ctx context.Context) error {
	return gw.server.Shutdown(ctx)
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
