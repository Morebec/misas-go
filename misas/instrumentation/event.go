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
	"github.com/morebec/misas-go/misas/event"
	"github.com/morebec/misas-go/misas/event/processing"
	"github.com/morebec/misas-go/misas/event/store"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"reflect"
)

// OpenTelemetryEventBusDecorator is a decorator allowing instrumenting an event.Bus.
type OpenTelemetryEventBusDecorator struct {
	event.Bus
	Tracer *SystemTracer
}

func (b *OpenTelemetryEventBusDecorator) Send(ctx context.Context, e event.Event) error {
	ctx, span := b.Tracer.Start(ctx, "eventBus.Send")
	defer span.End()

	span.SetAttributes(attribute.String("event.typeName", string(e.TypeName())))

	if err := b.Bus.Send(ctx, e); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

func (b *OpenTelemetryEventBusDecorator) RegisterHandler(t event.TypeName, h event.Handler) {
	b.Bus.RegisterHandler(t, func() event.HandlerFunc {
		return func(ctx context.Context, e event.Event) error {
			ctx, span := b.Tracer.Start(ctx, fmt.Sprintf("%s.%s", t, typeAsString(h)))
			defer span.End()
			if err := h.Handle(ctx, e); err != nil {
				span.RecordError(err)
				return err
			}

			return nil
		}
	}())
}

// typeAsString Returns the name of a type as a string.
func typeAsString(i any) string {
	t := reflect.TypeOf(i)
	if t.PkgPath() == "" && t.Name() == "" {
		// built-in type
		return t.String()
	}

	return fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())
}

type OpenTelemetryEventStoreDecorator struct {
	store.EventStore
	Tracer *SystemTracer
}

func (o *OpenTelemetryEventStoreDecorator) GlobalStreamID() store.StreamID {
	return o.EventStore.GlobalStreamID()
}

func (o *OpenTelemetryEventStoreDecorator) AppendToStream(ctx context.Context, streamID store.StreamID, events []store.EventDescriptor, opts ...store.AppendToStreamOption) error {
	ctx, span := o.Tracer.Start(ctx, "eventStore.AppendToStream")
	defer span.End()

	options := store.BuildAppendToStreamOptions(opts)

	span.SetAttributes(semconv.DBSystemKey.String("eventstore"))
	span.SetAttributes(semconv.DBStatementKey.String(string("AppendToStream " + streamID)))
	span.SetAttributes(semconv.DBOperationKey.String(string("AppendToStream " + streamID)))
	span.SetAttributes(attribute.String("db.eventstore.streamId", string(streamID)))

	var expectedVersion int
	if options.ExpectedVersion != nil {
		expectedVersion = int(*options.ExpectedVersion)
	} else {
		expectedVersion = int(store.InitialVersion)
	}
	span.SetAttributes(attribute.Int("db.statement.options.expectedVersion", expectedVersion))

	var typeNames []string
	for _, e := range events {
		typeNames = append(typeNames, string(e.TypeName))
	}
	span.SetAttributes(attribute.StringSlice("db.statement.options.eventTypeNames", typeNames))

	if err := o.EventStore.AppendToStream(ctx, streamID, events, opts...); err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (o *OpenTelemetryEventStoreDecorator) ReadFromStream(ctx context.Context, streamID store.StreamID, opts ...store.ReadFromStreamOption) (store.StreamSlice, error) {

	options := store.BuildReadFromStreamOptions(opts)

	ctx, span := o.Tracer.Start(ctx, "eventStore.ReadFromStream")
	defer span.End()

	span.SetAttributes(semconv.DBSystemKey.String("eventstore"))
	span.SetAttributes(semconv.DBStatementKey.String(string("ReadFromStream " + streamID)))
	span.SetAttributes(semconv.DBOperationKey.String(string("ReadFromStream " + streamID)))
	span.SetAttributes(attribute.String("db.eventstore.streamId", string(streamID)))
	span.SetAttributes(attribute.Int("db.statement.options.position", int(options.Position)))
	span.SetAttributes(attribute.Int("db.statement.options.maxCount", options.MaxCount))
	span.SetAttributes(attribute.String("db.statement.options.direction", string(options.Direction)))

	if options.EventTypeNameFilter != nil {
		span.SetAttributes(attribute.String("db.statement.options.filter.mode", string(options.EventTypeNameFilter.Mode)))

		var typeNames []string
		for _, typeName := range options.EventTypeNameFilter.EventTypeNames {
			typeNames = append(typeNames, string(typeName))
		}
		span.SetAttributes(attribute.StringSlice("db.statement.options.filter.eventTypeNames", typeNames))
	}

	stream, err := o.EventStore.ReadFromStream(ctx, streamID, opts...)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return store.StreamSlice{}, err
	}

	return stream, nil
}

func (o *OpenTelemetryEventStoreDecorator) TruncateStream(ctx context.Context, streamID store.StreamID, opts ...store.TruncateStreamOption) error {
	ctx, span := o.Tracer.Start(ctx, "eventStore.TruncateStream")
	defer span.End()

	options := store.BuildTruncateFromStreamOptions(opts)

	span.SetAttributes(semconv.DBSystemKey.String("eventstore"))
	span.SetAttributes(semconv.DBStatementKey.String(string("TruncateStream " + streamID)))
	span.SetAttributes(semconv.DBOperationKey.String(string("TruncateStream " + streamID)))
	span.SetAttributes(attribute.String("db.eventstore.streamId", string(streamID)))
	span.SetAttributes(attribute.Int("db.statement.options.beforePosition", int(options.BeforePosition)))

	if err := o.EventStore.TruncateStream(ctx, streamID, opts...); err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (o *OpenTelemetryEventStoreDecorator) DeleteStream(ctx context.Context, id store.StreamID) error {
	ctx, span := o.Tracer.Start(ctx, "eventStore.DeleteStream")
	defer span.End()

	span.SetAttributes(semconv.DBSystemKey.String("eventstore"))
	span.SetAttributes(semconv.DBStatementKey.String(string("DeleteStream " + id)))
	span.SetAttributes(semconv.DBOperationKey.String("DeleteStream"))
	span.SetAttributes(attribute.String("db.eventstore.streamId", string(id)))

	if err := o.EventStore.DeleteStream(ctx, id); err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (o *OpenTelemetryEventStoreDecorator) SubscribeToStream(ctx context.Context, streamID store.StreamID, opts ...store.SubscribeToStreamOption) (store.Subscription, error) {
	ctx, span := o.Tracer.Start(ctx, "eventStore.SubscribeToStream")
	defer span.End()

	options := store.BuildSubscribeToStreamOptions(opts)

	span.SetAttributes(semconv.DBSystemKey.String("eventstore"))
	span.SetAttributes(semconv.DBStatementKey.String(string("SubscribeToStream " + streamID)))
	span.SetAttributes(semconv.DBOperationKey.String("SubscribeToStream"))
	span.SetAttributes(attribute.String("db.eventstore.streamId", string(streamID)))

	if options.EventTypeNameFilter != nil {
		span.SetAttributes(attribute.String("db.statement.options.filter.mode", string(options.EventTypeNameFilter.Mode)))

		var typeNames []string
		for _, typeName := range options.EventTypeNameFilter.EventTypeNames {
			typeNames = append(typeNames, string(typeName))
		}
		span.SetAttributes(attribute.StringSlice("db.statement.options.filter.eventTypeNames", typeNames))
	}

	stream, err := o.EventStore.SubscribeToStream(ctx, streamID, opts...)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return store.Subscription{}, err
	}

	return stream, nil
}

func (o *OpenTelemetryEventStoreDecorator) StreamExists(ctx context.Context, id store.StreamID) (bool, error) {
	ctx, span := o.Tracer.Start(ctx, "eventStore.StreamExists")
	defer span.End()

	span.SetAttributes(semconv.DBSystemKey.String("eventstore"))
	span.SetAttributes(semconv.DBStatementKey.String(string("StreamExists " + id)))
	span.SetAttributes(semconv.DBOperationKey.String("StreamExists"))
	span.SetAttributes(attribute.String("db.eventstore.streamId", string(id)))

	exists, err := o.EventStore.StreamExists(ctx, id)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}

	return exists, nil
}

func (o *OpenTelemetryEventStoreDecorator) GetStream(ctx context.Context, id store.StreamID) (store.Stream, error) {
	ctx, span := o.Tracer.Start(ctx, "eventStore.GetStream")
	defer span.End()

	span.SetAttributes(semconv.DBSystemKey.String("eventstore"))
	span.SetAttributes(semconv.DBStatementKey.String(string("GetStream " + id)))
	span.SetAttributes(semconv.DBOperationKey.String("GetStream"))
	span.SetAttributes(attribute.String("db.eventstore.streamId", string(id)))

	stream, err := o.EventStore.GetStream(ctx, id)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return store.Stream{}, err
	}

	return stream, nil
}

func (o *OpenTelemetryEventStoreDecorator) Clear(ctx context.Context) error {
	ctx, span := o.Tracer.Start(ctx, "eventStore.Clear")
	defer span.End()

	span.SetAttributes(semconv.DBSystemKey.String("eventstore"))
	span.SetAttributes(semconv.DBStatementKey.String("Clear"))
	span.SetAttributes(semconv.DBOperationKey.String("Clear"))

	if err := o.EventStore.Clear(ctx); err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (o *OpenTelemetryEventStoreDecorator) Decorated() store.EventStore {
	return o.EventStore
}

type OpenTelemetryCheckpointStoreDecorator struct {
	processing.CheckpointStore
	Tracer SystemTracer
}

func (d *OpenTelemetryCheckpointStoreDecorator) Save(ctx context.Context, checkpoint processing.Checkpoint) error {
	ctx, span := d.Tracer.Start(ctx, "checkpointStore.Save")
	defer span.End()

	span.SetAttributes(semconv.DBSystemKey.String("checkpointstore"))
	span.SetAttributes(semconv.DBStatementKey.String(string("Save " + checkpoint.ID)))
	span.SetAttributes(semconv.DBOperationKey.String("Save"))

	if err := d.CheckpointStore.Save(ctx, checkpoint); err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (d *OpenTelemetryCheckpointStoreDecorator) FindById(ctx context.Context, id processing.CheckpointID) (*processing.Checkpoint, error) {
	ctx, span := d.Tracer.Start(ctx, "checkpointStore.FindById")
	defer span.End()

	span.SetAttributes(semconv.DBSystemKey.String("checkpointstore"))
	span.SetAttributes(semconv.DBStatementKey.String(string("FindById " + id)))
	span.SetAttributes(semconv.DBOperationKey.String("FindById"))

	checkpoint, err := d.CheckpointStore.FindById(ctx, id)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return checkpoint, nil
}

func (d *OpenTelemetryCheckpointStoreDecorator) Remove(ctx context.Context, id processing.CheckpointID) error {
	ctx, span := d.Tracer.Start(ctx, "checkpointStore.Remove")
	defer span.End()

	span.SetAttributes(semconv.DBSystemKey.String("checkpointstore"))
	span.SetAttributes(semconv.DBStatementKey.String(string("Remove " + id)))
	span.SetAttributes(semconv.DBOperationKey.String("Remove"))

	if err := d.CheckpointStore.Remove(ctx, id); err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}
