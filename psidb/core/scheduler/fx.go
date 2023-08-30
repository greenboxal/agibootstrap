package scheduler

import "go.uber.org/fx"

var Module = fx.Module(
	"core/scheduler",

	fx.Provide(NewScheduler),
	fx.Provide(NewSyncManager),
)
