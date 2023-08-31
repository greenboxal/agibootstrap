package scheduler

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
)

var logger = logging.GetLogger("scheduler")

var Module = fx.Module(
	"core/scheduler",

	fx.Provide(NewScheduler),
	fx.Provide(NewSyncManager),
)
