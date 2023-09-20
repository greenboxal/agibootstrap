package online

import (
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
)

var logger = logging.GetLogger("livegraph")
var tracer = otel.Tracer("livegraph", trace.WithInstrumentationAttributes(
	semconv.ServiceName("psidb-graph"),
))
