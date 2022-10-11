// Copyright 2022 Morébec
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package system

import (
	"github.com/morebec/go-system/misas/instrumentation"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

type InstrumentationOption func(s *System)

func WithInstrumentation(opts ...InstrumentationOption) Option {
	return func(s *System) {
		for _, opt := range opts {
			opt(s)
		}
	}
}

func WithDefaultTracer() InstrumentationOption {
	return WithTracer(instrumentation.NewSystemTracer())
}

func WithTracer(tracer *instrumentation.SystemTracer) InstrumentationOption {
	return func(s *System) {
		s.Tracer = tracer
	}
}

func WithTracingSpanExporter(se trace.SpanExporter) Option {
	return func(s *System) {
		s.SpanExporter = se
	}
}

func WithLogger(logger *otelzap.Logger) InstrumentationOption {
	return func(s *System) {
		s.Logger = logger
	}
}

func WithDefaultLogger() InstrumentationOption {
	return func(s *System) {
		var logger *zap.Logger
		var err error
		switch s.Environment {
		case Dev, Test, Staging:
			logger, err = zap.NewDevelopment()
		case Prod:
			logger, err = zap.NewProduction()
		}
		if err != nil {
			panic(err)
		}
		WithLogger(otelzap.New(logger))(s)
	}
}

func WithJaegerTracingSpanExporter(url string) InstrumentationOption {
	return func(s *System) {
		exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
		if err != nil {
			panic(err)
		}
		WithTracingSpanExporter(exporter)(s)
	}
}

func WithCommandBusInstrumentation() InstrumentationOption {
	return func(s *System) {
		s.CommandBus = &instrumentation.OpenTelemetryCommandBusDecorator{
			Bus:    s.CommandBus,
			Tracer: s.Tracer,
		}
	}
}

func WithQueryBusInstrumentation() InstrumentationOption {
	return func(s *System) {
		s.QueryBus = &instrumentation.OpenTelemetryQueryBusDecorator{
			Bus:    s.QueryBus,
			Tracer: s.Tracer,
		}
	}
}

func WithEventBusInstrumentation() InstrumentationOption {
	return func(s *System) {
		s.EventBus = &instrumentation.OpenTelemetryEventBusDecorator{
			Bus:    s.EventBus,
			Tracer: s.Tracer,
		}
	}
}

func WithEventStoreInstrumentation() InstrumentationOption {
	return func(s *System) {
		s.EventStore = &instrumentation.OpenTelemetryEventStoreDecorator{
			EventStore: s.EventStore,
			Tracer:     s.Tracer,
		}
	}
}

func WithPredictionBusInstrumentation() InstrumentationOption {
	return func(s *System) {
		s.PredictionBus = &instrumentation.OpenTelemetryPredictionBusDecorator{
			Bus:    s.PredictionBus,
			Tracer: s.Tracer,
		}
	}
}

// WithEntryPointInstrumentation Indicates that the entry point should have instrumentation enabled on it.
func WithEntryPointInstrumentation() EntryPointOption {
	return func(e EntryPoint) {
		e = NewTracingEntryPointDecorator(e)
	}
}
