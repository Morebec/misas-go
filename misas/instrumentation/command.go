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
	"github.com/jwillp/go-system/misas/command"
	"go.opentelemetry.io/otel/attribute"
)

// OpenTelemetryCommandBusDecorator decorator allowing instrumenting a command.Bus.
type OpenTelemetryCommandBusDecorator struct {
	command.Bus
	Tracer *SystemTracer
}

func (b *OpenTelemetryCommandBusDecorator) RegisterHandler(t command.TypeName, h command.Handler) {
	b.Bus.RegisterHandler(t, func() command.HandlerFunc {
		return func(ctx context.Context, c command.Command) (any, error) {
			ctx, span := b.Tracer.Start(ctx, fmt.Sprintf("%s.handle", t))
			defer span.End()

			span.SetAttributes(attribute.String("command.typeName", string(t)))

			events, err := h.Handle(ctx, c)
			if err != nil {
				span.RecordError(err)
				return nil, err
			}

			return events, nil
		}
	}())
}

func (b *OpenTelemetryCommandBusDecorator) Send(ctx context.Context, c command.Command) (any, error) {
	ctx, span := b.Tracer.Start(ctx, "commandBus.Send")
	defer span.End()

	span.SetAttributes(attribute.String("command.typeName", string(c.TypeName())))

	fulfillmentResult, err := b.Bus.Send(ctx, c)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return fulfillmentResult, nil
}
