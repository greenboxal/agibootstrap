package apimachinery

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/riandyrn/otelchi"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"moul.io/chizap"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type Server struct {
	logger *otelzap.SugaredLogger
	server http.Server
	mux    chi.Router
	cfg    *coreapi.Config
}

func NewServer(
	lc fx.Lifecycle,
	cfg *coreapi.Config,
	sm coreapi.SessionManager,
) *Server {
	api := &Server{}

	api.logger = logging.GetLogger("api")
	api.cfg = cfg
	api.mux = chi.NewRouter()
	api.server.Handler = api.mux

	api.mux.Use(otelchi.Middleware("psidb", otelchi.WithChiRoutes(api.mux)))
	api.mux.Use(middleware.RealIP)
	api.mux.Use(middleware.RequestID)
	//api.mux.Use(NewRequestLogger(logging.GetLogger("apimachinery").Desugar(), &chizap.Opts{}))
	api.mux.Use(middleware.Recoverer)

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"*", "X-PsiDB-Session-ID"},
		AllowCredentials: true,
	})

	api.mux.Use(corsHandler.Handler)

	api.mux.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			sessionId := request.Header.Get("X-PsiDB-Session-ID")

			if sessionId != "" {
				sess := sm.GetOrCreateSession(coreapi.SessionConfig{
					SessionID: sessionId,
				})

				if sess != nil {
					ctx := request.Context()
					ctx = coreapi.WithSession(ctx, sess)
					request = request.WithContext(ctx)

					sess.KeepAlive()
				}
			}

			next.ServeHTTP(writer, request)
		})
	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return api.Start(ctx)
		},

		OnStop: func(ctx context.Context) error {
			return api.Stop(ctx)
		},
	})

	return api
}

func (a *Server) Mount(path string, handler http.Handler) {
	a.mux.Mount(path, http.StripPrefix(path, handler))
}

func (a *Server) Start(ctx context.Context) error {
	var listener net.Listener

	endpoint := a.cfg.ListenEndpoint

	if endpoint == "" {
		endpoint = "0.0.0.0:22440"
	}

	if a.cfg.UseTLS {
		tlsConfig := &tls.Config{}

		a.server.TLSConfig = tlsConfig

		if a.cfg.TLSCertFile != "" && a.cfg.TLSKeyFile != "" {
			cert, err := tls.LoadX509KeyPair(a.cfg.TLSCertFile, a.cfg.TLSKeyFile)

			if err != nil {
				return fmt.Errorf("failed to load TLS certificate: %w", err)
			}

			tlsConfig.Certificates = []tls.Certificate{cert}
		} else {
			serverTLSCert, err := generateSelfSigned(a.cfg)

			if err != nil {
				return fmt.Errorf("failed to generate self signed certificate: %w", err)
			}

			tlsConfig.Certificates = []tls.Certificate{serverTLSCert}
		}

		l, err := tls.Listen("tcp", endpoint, tlsConfig)

		if err != nil {
			return err
		}

		listener = l
	} else {
		l, err := net.Listen("tcp", endpoint)

		if err != nil {
			return err
		}

		listener = l
	}

	a.logger.Infow("Server is listening", "endpoint", endpoint)

	go func() {
		if err := a.server.Serve(listener); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				a.logger.Error(err)
			}
		}
	}()

	return nil
}

func (a *Server) Stop(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}

func generateSelfSigned(cfg *coreapi.Config) (tls.Certificate, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		panic(err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 180),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := []string{"localhost", "127.0.0.1", "::1", "0.0.0.0"}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)

	if err != nil {
		panic(err)
	}

	out := &bytes.Buffer{}

	if err := pem.Encode(out, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		panic(err)
	}

	if err := pem.Encode(out, pemBlockForKey(priv)); err != nil {
		panic(err)
	}

	return tls.X509KeyPair(out.Bytes(), out.Bytes())
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}

func NewRequestLogger(logger *otelzap.Logger, opts *chizap.Opts) func(next http.Handler) http.Handler {
	if logger == nil {
		return func(next http.Handler) http.Handler { return next }
	}
	if opts == nil {
		opts = &chizap.Opts{}
	}
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()
			defer func() {
				reqLogger := logger.With(
					zap.String("proto", r.Proto),
					zap.String("path", r.URL.Path),
					zap.String("reqId", middleware.GetReqID(r.Context())),
					zap.Duration("lat", time.Since(t1)),
					zap.Int("status", ww.Status()),
					zap.Int("size", ww.BytesWritten()),
				)
				if opts.WithReferer {
					ref := ww.Header().Get("Referer")
					if ref == "" {
						ref = r.Header.Get("Referer")
					}
					if ref != "" {
						reqLogger = reqLogger.With(zap.String("ref", ref))
					}
				}
				if opts.WithUserAgent {
					ua := ww.Header().Get("User-Agent")
					if ua == "" {
						ua = r.Header.Get("User-Agent")
					}
					if ua != "" {
						reqLogger = reqLogger.With(zap.String("ua", ua))
					}
				}
				ctxLogger := otelzap.New(reqLogger)
				ctxLogger.Info("Served")
			}()
			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
