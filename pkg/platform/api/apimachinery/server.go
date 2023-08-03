package apimachinery

import (
	"context"
	"net"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Server struct {
	logger *zap.SugaredLogger
	server http.Server
	mux    chi.Router
}

func NewServer(
	lc fx.Lifecycle,
	logger *zap.SugaredLogger,
) *Server {
	api := &Server{}

	api.logger = logger.Named("api")
	api.mux = chi.NewRouter()
	api.server.Handler = api.mux

	api.mux.Use(middleware.RealIP)
	api.mux.Use(middleware.RequestID)
	api.mux.Use(middleware.Logger)
	api.mux.Use(middleware.Recoverer)

	api.mux.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return api.Start(ctx)
		},

		OnStop: func(ctx context.Context) error {
			return api.Shutdown(ctx)
		},
	})

	return api
}

func (a *Server) Mount(path string, handler http.Handler) {
	a.mux.Mount(path, http.StripPrefix(path, handler))
}

func (a *Server) Start(ctx context.Context) error {
	endpoint := os.Getenv("AGIB_LISTEN_ENDPOINT")

	if endpoint == "" {
		endpoint = "0.0.0.0:22440"
	}

	l, err := net.Listen("tcp", endpoint)

	if err != nil {
		return err
	}

	a.logger.Infow("Server is listening", "endpoint", endpoint)

	go func() {
		if err := a.server.Serve(l); err != nil {
			if err != http.ErrServerClosed {
				a.logger.Error(err)
			}
		}
	}()

	return nil
}

func (a *Server) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}
