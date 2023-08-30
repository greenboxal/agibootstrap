package agents

import (
	"go.opentelemetry.io/otel"
	"go.uber.org/fx"
)

var tracer = otel.Tracer("agents")

var Module = fx.Module(
	"services/agents",
)
