package main

import (
	"os"
	"path"

	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/api/apimachinery"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/workspace"
	"github.com/greenboxal/agibootstrap/psidb"
	"github.com/greenboxal/agibootstrap/psidb/config"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/network"
)

func main() {
	logging.Initialize()
	defer logging.Shutdown()

	wd, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	app := fx.New(
		// FIXME
		fx.Provide(NewDefaultNodeEmbedder),

		psidb.BaseModules,
		workspace.Module,

		fx.Provide(func(c *config.Config, lrm config.LocalResourceManager) *coreapi.Config {
			cfg := &coreapi.Config{}
			cfg.ListenEndpoint = lrm.ListenEndpoint("api").String()
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

		fx.Invoke(func(core coreapi.Core) {}),
		fx.Invoke(func(server *network.Network) {}),
		fx.Invoke(func(server *apimachinery.Server) {}),
		//fx.Invoke(func(server *fuse.Manager) {}),
	)

	app.Run()
}
