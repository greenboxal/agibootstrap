package core

import (
	"github.com/ipld/go-ipld-prime/linking"
	"go.uber.org/dig"
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/db/indexing"
)

var Module = fx.Module(
	"core",

	fx.Provide(inject.NewServiceProvider),

	fx.Provide(NewCheckpoint),
	fx.Provide(NewJournal),
	fx.Provide(NewDataStore),
	fx.Provide(NewIndexManager),
	fx.Provide(NewBlockManager),
	fx.Provide(NewSessionManager),
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
		im *indexing.Manager,
		vfsm *vfs.Manager,
	) {
		inject.RegisterInstance[coreapi.Core](core.sp, core)
		inject.RegisterInstance[*coreapi.Config](core.sp, core.cfg)
		inject.RegisterInstance[*indexing.Manager](core.sp, im)
		inject.RegisterInstance[*linking.LinkSystem](core.sp, &core.lsys)
		inject.RegisterInstance[*vfs.Manager](core.sp, vfsm)
	}),
)
