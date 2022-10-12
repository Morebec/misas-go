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

package domain

import (
	"context"
	"github.com/google/uuid"
	"github.com/morebec/misas-go/misas"
	"github.com/morebec/misas-go/misas/event"
	"github.com/morebec/misas-go/misas/event/store"
	"github.com/pkg/errors"
)

// Version Represents the Version of an EventSourcedAggregate for concurrency checking.
type Version store.StreamVersion

// Incremented Returns the value of this version incremented by one.
func (v Version) Incremented() Version {
	return v + 1
}

// EventSourcedAggregate Represents a consistency boundary of a given concept in a system.
// In the context of event sourcing, an aggregate is both a projection and a change agent.
// Internally an EventSourcedAggregate should save its state changes as a series of recorded events.
type EventSourcedAggregate interface {
	// ApplyEvent applies an event to this aggregate without recording it as a recorded change.
	ApplyEvent(e event.Event)

	// RecordedEvents returns the list of events recorded as state changes by this aggregate.
	RecordedEvents() event.List

	// ClearRecordedEvents clears the list of recorded events on this aggregate.
	ClearRecordedEvents()
}

// EventSourcedAggregateBase Represents an embeddable struct that implements the event recording methods of the EventSourcedAggregate interface.
// It takes a applyEvent callback in order to apply events to a given aggregate when an event is recorded.
type EventSourcedAggregateBase struct {
	ApplyEvent     func(evt event.Event)
	recordedEvents event.List
}

func (r *EventSourcedAggregateBase) RecordedEvents() event.List {
	return r.recordedEvents
}

func (r *EventSourcedAggregateBase) ClearRecordedEvents() {
	r.recordedEvents = nil
}

// RecordEvent records an event and applies it as per its applyEvent callback.
func (r *EventSourcedAggregateBase) RecordEvent(e event.Event) {
	r.recordedEvents = append(r.recordedEvents, e)
	if r.ApplyEvent == nil {
		panic("no apply event callback has been provided.")
	}
	r.ApplyEvent(e)
}

// EventStoreRepository is a service responsible for easily storing the recorded events of EventSourcedAggregate into an store.EventStore.
type EventStoreRepository[T EventSourcedAggregate] struct {
	eventStore       store.EventStore
	eventConverter   *store.EventConverter
	newAggregateFunc func() T
}

func NewEventStoreRepository[T EventSourcedAggregate](eventStore store.EventStore, eventConverter *store.EventConverter) EventStoreRepository[T] {
	return EventStoreRepository[T]{eventStore: eventStore, eventConverter: eventConverter}
}

func (r EventStoreRepository[T]) Add(ctx context.Context, streamId store.StreamID, a T) error {
	return r.Update(ctx, streamId, a, Version(store.InitialVersion))
}

func (r EventStoreRepository[T]) Update(ctx context.Context, streamId store.StreamID, a T, version Version) error {
	expectedVersion := store.StreamVersion(version)
	var descriptors []store.EventDescriptor

	for _, e := range a.RecordedEvents() {
		payload, err := r.eventConverter.ToEventPayload(e)
		if err != nil {
			return errors.Wrapf(err, "failed saving aggregate to stream \"%s\"", streamId)
		}
		m := misas.Metadata{}

		descriptors = append(descriptors, store.EventDescriptor{
			ID:       store.EventID(uuid.New().String()),
			TypeName: e.TypeName(),
			Payload:  payload,
			Metadata: m,
		})
	}

	if err := r.eventStore.AppendToStream(ctx, streamId, descriptors, store.WithExpectedVersion(expectedVersion)); err != nil {
		return errors.Wrapf(err, "failed saving aggregate to stream \"%s\"", streamId)
	}

	a.ClearRecordedEvents()
	return nil
}

func (r EventStoreRepository[T]) Load(ctx context.Context, id store.StreamID) (T, Version, error) {
	stream, err := r.eventStore.ReadFromStream(ctx, id, store.FromStart(), store.InForwardDirection())
	if err != nil {
		var out T
		return out, 0, errors.Wrapf(err, "failed loading aggregate from stream \"%s\"", id)
	}

	var aggregate T
	version := Version(store.InitialVersion)

	for _, d := range stream.Descriptors {
		e, err := r.eventConverter.FromRecordedEventDescriptor(d)
		if err != nil {
			var out T
			return out, 0, errors.Wrapf(err, "failed loading aggregate from stream \"%s\"", id)
		}
		aggregate.ApplyEvent(e)
		version = version.Incremented()
	}

	return aggregate, version, nil
}
