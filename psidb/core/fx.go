package core

import (
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/dig"
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/core/scheduler"
	"github.com/greenboxal/agibootstrap/psidb/core/session"
	"github.com/greenboxal/agibootstrap/psidb/core/vfs"
	"github.com/greenboxal/agibootstrap/psidb/core/vm"
)

var tracer = otel.Tracer("psidb", trace.WithInstrumentationAttributes(
	semconv.ServiceName("psidb-core"),
))

var Module = fx.Module(
	"core",

	session.Module,
	scheduler.Module,
	vfs.Module,
	vm.FXModule,

	fx.Provide(NewCheckpoint),
	fx.Provide(NewJournal),
	fx.Provide(NewDataStore),
	fx.Provide(NewMetadataStore),
	fx.Provide(NewBlockManager),
	fx.Provide(NewDispatcher),
	fx.Provide(NewCore),

	fx.Provide(func(core *Core) (res struct {
		dig.Out

		Core           coreapi.Core
		ServiceLocator inject.ServiceLocator
	}) {
		res.Core = core
		res.ServiceLocator = core.serviceProvider

		return
	}),

	inject.WithRegisteredService[coreapi.Core](inject.ServiceRegistrationScopeSingleton),
	inject.WithRegisteredService[*coreapi.Config](inject.ServiceRegistrationScopeSingleton),
)
