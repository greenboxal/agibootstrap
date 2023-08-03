package core

import (
	"go.uber.org/dig"
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/db/indexing"
)

var Module = fx.Module(
	"core",

	fx.Provide(inject.NewServiceProvider),

	fx.Provide(NewCheckpoint),
	fx.Provide(NewJournal),
	fx.Provide(NewDataStore),
	fx.Provide(NewBlockManager),
	fx.Provide(NewCore),

	fx.Provide(func(core *Core) (res struct {
		dig.Out

		Core           coreapi.Core
		ServiceLocator inject.ServiceLocator
		IndexManager   *indexing.Manager
	}) {
		res.Core = core
		res.ServiceLocator = core.sp
		res.IndexManager = core.indexManager

		return
	}),
)
