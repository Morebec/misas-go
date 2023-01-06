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
	"github.com/morebec/misas-go/misas/query"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// OpenTelemetryQueryBusDecorator is a decorator allowing instrumenting a query.Bus.
type OpenTelemetryQueryBusDecorator struct {
	query.Bus
	Tracer *SystemTracer
}

func (b *OpenTelemetryQueryBusDecorator) RegisterHandler(t query.PayloadTypeName, h query.Handler) {
	b.Bus.RegisterHandler(t, func() query.HandlerFunc {
		return func(ctx context.Context, q query.Query) (any, error) {
			ctx, span := b.Tracer.Start(ctx, fmt.Sprintf("%s.handle", t))
			defer span.End()

			data, err := h.Handle(ctx, q)
			if err != nil {
				span.RecordError(err, trace.WithStackTrace(true))
				span.SetStatus(codes.Error, err.Error())
				return nil, err
			}

			return data, nil
		}
	}())
}

func (b *OpenTelemetryQueryBusDecorator) Send(ctx context.Context, q query.Query) (any, error) {
	ctx, span := b.Tracer.Start(ctx, "queryBus.Send")
	defer span.End()

	span.SetAttributes(attribute.String("query.typeName", string(q.Payload.TypeName())))

	data, err := b.Bus.Send(ctx, q)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return data, nil
}
