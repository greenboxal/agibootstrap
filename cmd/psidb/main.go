package main

import (
	"os"
	"path"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/greenboxal/agibootstrap/pkg/platform/api/apimachinery"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/workspace"
	"github.com/greenboxal/agibootstrap/psidb"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

func main() {
	logging.Initialize()
	defer logging.Shutdown()

	wd, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	app := fx.New(
		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.ZapLogger{Logger: logging.GetRootLogger().Logger}
		}),

		// FIXME
		fx.Provide(NewDefaultNodeEmbedder),

		psidb.BaseModules,
		workspace.Module,

		fx.Provide(func() *coreapi.Config {
			cfg := &coreapi.Config{}
			cfg.RootUUID = "QmYXZ"
			cfg.ProjectDir = wd
			cfg.DataDir = path.Join(cfg.ProjectDir, ".fti/psi")
			cfg.UseTLS = true
			cfg.TLSCertFile = os.Getenv("PSIDB_TLS_CERT_FILE")
			cfg.TLSKeyFile = os.Getenv("PSIDB_TLS_KEY_FILE")

			cfg.Workers.MaxWorkers = 32
			cfg.Workers.MaxCapacity = 1024

			cfg.SetDefaults()

			return cfg
		}),

		fx.Invoke(func(server *apimachinery.Server) {}),

		fx.Invoke(func(wrk *workspace.Workspace) {
		}),
	)

	app.Run()
}
