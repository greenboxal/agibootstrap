package tracing

import (
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	ddotel "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentelemetry"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
)

var providerInstance *ddotel.TracerProvider
var logger = logging.GetLogger("tracing")

func Initialize() {
	if os.Getenv("PSIDB_ENABLE_OTEL") != "" {
		if providerInstance == nil {
			providerInstance = ddotel.NewTracerProvider(
				tracer.WithService("psidb"),
				tracer.WithEnv("dev"),
				tracer.WithProfilerEndpoints(true),
				tracer.WithProfilerCodeHotspots(true),
			)

			otel.SetTracerProvider(providerInstance)
			otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
		}

		err := profiler.Start(
			profiler.WithService("psidb"),
			profiler.WithEnv("dev"),
			profiler.WithVersion("0.1"),
			profiler.WithProfileTypes(
				profiler.CPUProfile,
				profiler.HeapProfile,
				profiler.BlockProfile,
				profiler.MutexProfile,
				profiler.GoroutineProfile,
			),
		)

		if err != nil {
			log.Fatal(err)
		}

	}
}

func Shutdown() {
	if providerInstance == nil {
		return
	}

	profiler.Stop()

	if err := providerInstance.Shutdown(); err != nil {
		logger.Error(err)
	}
}
