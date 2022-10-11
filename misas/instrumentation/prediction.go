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
	"github.com/morebec/go-system/misas/prediction"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// OpenTelemetryPredictionBusDecorator is a decorator allowing to instrument a prediction.Bus.
type OpenTelemetryPredictionBusDecorator struct {
	prediction.Bus
	Tracer *SystemTracer
}

func (o *OpenTelemetryPredictionBusDecorator) Send(ctx context.Context, p prediction.Prediction) error {
	ctx, span := o.Tracer.Start(ctx, "predictionBus.Send")
	defer span.End()

	span.SetAttributes(attribute.String("prediction.typeName", string(p.TypeName())))

	if err := o.Bus.Send(ctx, p); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (o *OpenTelemetryPredictionBusDecorator) RegisterHandler(t prediction.TypeName, h prediction.Handler) {
	o.Bus.RegisterHandler(t, func() prediction.HandlerFunc {
		return func(p prediction.Prediction, ctx context.Context) error {
			ctx, span := o.Tracer.Start(ctx, fmt.Sprintf("%s.%s", t, typeAsString(h)))
			defer span.End()
			if err := h.Handle(p, ctx); err != nil {
				span.RecordError(err, trace.WithStackTrace(true))
				span.SetStatus(codes.Error, err.Error())
				return err
			}

			return nil
		}
	}())
}
