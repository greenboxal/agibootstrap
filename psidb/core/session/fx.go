package session

import (
	`go.opentelemetry.io/otel`
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	`github.com/greenboxal/agibootstrap/pkg/platform/logging`
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

var logger = logging.GetLogger("session")
var tracer = otel.Tracer("SessionManager")

var Module = fx.Module(
	"core/session",

	fx.Provide(NewManager),

	fx.Invoke(func(sp inject.ServiceProvider, m coreapi.SessionManager) {
		inject.RegisterInstance(sp, m)
	}),
)
