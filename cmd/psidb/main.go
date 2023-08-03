package main

import (
	"os"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/greenboxal/agibootstrap/pkg/platform/api/apimachinery"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/psidb"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

func main() {
	wd, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	app := fx.New(
		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.ZapLogger{Logger: logging.GetRootLogger()}
		}),

		psidb.BaseModules,

		fx.Provide(func() *coreapi.Config {
			cfg := &coreapi.Config{}
			cfg.DataDir = "./.fti/psi"
			cfg.RootUUID = "QmYXZ"
			cfg.ProjectDir = wd

			return cfg
		}),

		fx.Invoke(func(server *apimachinery.Server) {}),
	)

	app.Run()
}
