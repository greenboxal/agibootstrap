package main

import (
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/greenboxal/agibootstrap/pkg/platform/api/apimachinery"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/psidb/apis/rest"
	"github.com/greenboxal/agibootstrap/psidb/apis/rpc"
	rpcv1 "github.com/greenboxal/agibootstrap/psidb/apis/rpc/v1"
	"github.com/greenboxal/agibootstrap/psidb/core"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/modules"
)

var BaseModules = fx.Options(
	logging.Module,
	apimachinery.Module,
	core.Module,
	modules.Module,
	rest.Module,
	rpc.Module,
	rpcv1.Module,
)

func main() {
	app := fx.New(
		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.ZapLogger{Logger: logging.GetRootLogger()}
		}),

		BaseModules,

		fx.Provide(func() *coreapi.Config {
			cfg := &coreapi.Config{}
			cfg.DataDir = "./.fti/psi"
			cfg.RootUUID = "QmYXZ"

			return cfg
		}),

		fx.Invoke(func(server *apimachinery.Server) {}),
	)

	app.Run()
}
