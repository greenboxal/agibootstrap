package core

import (
	"github.com/ipld/go-ipld-prime/linking"
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
	"github.com/greenboxal/agibootstrap/psidb/services/indexing"
	"github.com/greenboxal/agibootstrap/psidb/services/typing"
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

	fx.Provide(inject.NewServiceProvider),

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
		res.ServiceLocator = core.sp

		return
	}),

	fx.Invoke(func(
		core *Core,
		disp *Dispatcher,
		sch *scheduler.Scheduler,
		im *indexing.Manager,
		vfsm *vfs.Manager,
		tm *typing.Manager,
		vms *vm.VM,
	) {
		inject.RegisterInstance[coreapi.Core](core.sp, core)
		inject.RegisterInstance[*coreapi.Config](core.sp, core.cfg)
		inject.RegisterInstance[*indexing.Manager](core.sp, im)
		inject.RegisterInstance[*linking.LinkSystem](core.sp, &core.lsys)
		inject.RegisterInstance[*vfs.Manager](core.sp, vfsm)
		inject.RegisterInstance[*typing.Manager](core.sp, tm)
		inject.RegisterInstance[*vm.VM](core.sp, vms)
	}),
)
