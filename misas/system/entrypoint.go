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
	"context"
	"fmt"
	"github.com/morebec/go-system/misas/event/processing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	openTrace "go.opentelemetry.io/otel/trace"
)

// EntryPoint corresponds to a system responsible for starting a specific point of the application.
// E.g. Web Server could be an entry point, while a projector could be a different entry point.
type EntryPoint interface {
	// Name Returns the name of the entry point. This can be used while logging to know which entry point encountered an error.
	Name() string

	// Start the entry point's execution.
	Start(ctx context.Context, s *System) error

	// Stop the entry point's execution.
	Stop(ctx context.Context, s *System) error
}

// default private implementation of an entry point.
type entryPoint struct {
	name  string
	start func(ctx context.Context, s *System) error
	stop  func(ctx context.Context, s *System) error
}

// NewEntryPoint allows creating a new entry point easily without the need to define a type.
func NewEntryPoint(
	name string,
	start func(ctx context.Context, s *System) error,
	stop func(ctx context.Context, s *System) error,
	opts ...EntryPointOption,
) *entryPoint {
	e := entryPoint{
		name:  name,
		start: start,
		stop:  stop,
	}

	for _, opt := range opts {
		opt(e)
	}

	return &e
}

func (e entryPoint) Name() string {
	return e.name
}

func (e entryPoint) Start(ctx context.Context, s *System) error {
	return e.start(ctx, s)
}

func (e entryPoint) Stop(ctx context.Context, s *System) error {
	return e.stop(ctx, s)
}

// NewTracingEntryPointDecorator sets up a decorator around an entry point to allow tracing.
func NewTracingEntryPointDecorator(entryPoint EntryPoint) EntryPoint {

	type ShutdownTracingFn func(ctx context.Context) error

	setupTracing := func(s *System) (ShutdownTracingFn, error) {
		r, _ := resource.Merge(
			resource.Default(),
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(s.Information.Name),
				semconv.ServiceVersionKey.String(s.Information.Version),
				attribute.String("entryPoint", entryPoint.Name()),
				semconv.DeploymentEnvironmentKey.String(string(s.Environment)),
			),
		)

		tp := trace.NewTracerProvider(
			trace.WithBatcher(s.SpanExporter),
			trace.WithSampler(trace.AlwaysSample()),
			trace.WithResource(r),
		)
		otel.SetTracerProvider(tp)

		otel.SetTextMapPropagator(propagation.TraceContext{})

		return tp.Shutdown, nil
	}

	var shutdownTracing ShutdownTracingFn
	var entryPointCtx context.Context
	var entryPointSpan openTrace.Span

	start := func(ctx context.Context, s *System) error {
		if shutdown, err := setupTracing(s); err != nil {
			s.Logger.Error(fmt.Sprintf("failed setting up tracing: %s", err.Error()))
			return err
		} else {
			shutdownTracing = shutdown
		}

		entryPointCtx, entryPointSpan = s.Tracer.Start(ctx, fmt.Sprintf("%s", entryPoint.Name()))

		ctx, span := s.Tracer.Start(entryPointCtx, fmt.Sprintf("%s.Start", entryPoint.Name()))
		defer span.End()

		if err := entryPoint.Start(ctx, s); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}

		return nil
	}

	stop := func(ctx context.Context, s *System) error {
		ctx, span := s.Tracer.Start(entryPointCtx, fmt.Sprintf("%s.Stop", entryPoint.Name()))
		defer func(ctx context.Context) {
			span.End()
			entryPointSpan.End()
			if err := shutdownTracing(ctx); err != nil {
				entryPointSpan.RecordError(err)
				entryPointSpan.SetStatus(codes.Error, err.Error())
				entryPointSpan.End()
				s.Logger.Error(fmt.Sprintf("failed shutting tracer provider down: %s", err.Error()))
			}
		}(ctx)

		if err := entryPoint.Stop(ctx, s); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}

		return nil
	}

	return NewEntryPoint(entryPoint.Name(), start, stop)
}

type EntryPointOption func(e EntryPoint)

func NewEventProcessorEntryPoint(name string, p processing.Processor) *entryPoint {
	return NewEntryPoint(name, func(ctx context.Context, s *System) error {
		return p.Run(ctx)
	}, func(ctx context.Context, s *System) error {
		return nil
	})
}
