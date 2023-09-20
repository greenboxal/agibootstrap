package agents

import (
	"go.opentelemetry.io/otel"
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
)

var tracer = otel.Tracer("agents")

var logger = logging.GetLogger("psi/mod/chat")

var Module = fx.Module(
	"services/agents",
)
