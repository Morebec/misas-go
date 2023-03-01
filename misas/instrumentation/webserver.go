package instrumentation

import (
	"github.com/morebec/misas-go/misas/httpapi"
	"github.com/riandyrn/otelchi"
	"go.opentelemetry.io/otel/trace"
)

type tracerProviderFunc func() trace.Tracer

func (t tracerProviderFunc) Tracer(name string, options ...trace.TracerOption) trace.Tracer {
	return t()
}

func EnableOpenTelemetryOnWebServer(tracer *SystemTracer, serverName string) httpapi.ServerOption {
	return func(w *httpapi.Server) {
		w.Router.Use(otelchi.Middleware(
			serverName,
			otelchi.WithChiRoutes(w.Router),
			otelchi.WithRequestMethodInSpanName(true),
			otelchi.WithTracerProvider(tracerProviderFunc(func() trace.Tracer {
				return tracer
			})),
		))
	}
}
