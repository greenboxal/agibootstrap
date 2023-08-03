package codex

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/psidb"
)

var FxModule = fx.Module(
	"codex",

	psidb.BaseModules,

	fx.Provide(NewPlatform),
	fx.Provide(NewSyncManager),
)
