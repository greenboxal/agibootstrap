package pubsub

import (
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

var tracer = otel.Tracer("pubsub", trace.WithInstrumentationAttributes(
	semconv.ServiceName("psidb-graph"),
	semconv.MessagingSystem("psidb-pubsub"),
))

var Module = fx.Module(
	"services/pubsub",

	fx.Provide(NewManager),
)
