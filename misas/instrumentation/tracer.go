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

package instrumentation

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	openTrace "go.opentelemetry.io/otel/trace"
)

// SystemTracer Tracer Wrapper around open telemetry Tracer.Tracer
type SystemTracer struct{}

func NewSystemTracer() *SystemTracer {
	return &SystemTracer{}
}

func (t SystemTracer) Start(ctx context.Context, spanName string, opts ...openTrace.SpanStartOption) (context.Context, openTrace.Span) {
	ctx, span := otel.Tracer("").Start(ctx, spanName, opts...)
	return ctx, SpanErrorStackDecorator{Span: span}
}

// Instrument allows instrumenting a certain function with a given span name.
func (t SystemTracer) Instrument(ctx context.Context, spanName string, f func(ctx context.Context) error, opts ...openTrace.SpanStartOption) error {
	ctx, span := t.Start(ctx, spanName, opts...)
	defer span.End()

	err := f(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

// SpanErrorStackDecorator Decorator of a span that allows saving the stacktrace of an error (when available) when using the method RecordError.
// It also deviates slightly from the openTrace.Span interface, where it will automatically set the status of the span as error.
// The out-of-the-box option openTrace.WithStackTrace does not check the error and simply dumps the stack trace where it happens.
// This implementation aims to correct this situation by trying to find the stack trace of the error instead.
// TL;DR: It serves to replace the default openTrace.WithStackTrace(true) with a stack trace that is inferred from the error if using the errors package
type SpanErrorStackDecorator struct {
	openTrace.Span
}

func (s SpanErrorStackDecorator) RecordError(err error, opts ...openTrace.EventOption) {

	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

	s.Span.SetStatus(codes.Error, err.Error())

	if tracer, ok := err.(stackTracer); ok {
		opts = append(opts, openTrace.WithAttributes(
			semconv.ExceptionStacktraceKey.String(fmt.Sprintf("%+v", tracer.StackTrace()))),
		)
	} else {
		opts = append(opts, openTrace.WithStackTrace(true))
	}

	s.Span.RecordError(err, opts...)
}
