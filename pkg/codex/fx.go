package codex

import "go.uber.org/fx"

var FxModule = fx.Module(
	"codex",

	fx.Provide(NewSyncManager),
)
