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
	"github.com/morebec/misas-go/misas/command"
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

// EventStoreRepository is a service responsible for easily storing and retrieving event.Event into a store.EventStore.
type EventStoreRepository struct {
	eventStore       store.EventStore
	eventConverter   *store.EventConverter
	metadataProvider func(e event.Event) misas.Metadata
	streamPrefix     string
}

func NewEventStoreRepository(
	eventStore store.EventStore,
	eventConverter *store.EventConverter,
	metadataProvider func(e event.Event) misas.Metadata,
) EventStoreRepository {
	if metadataProvider == nil {
		metadataProvider = func(e event.Event) misas.Metadata {
			return misas.Metadata{}
		}
	}

	return EventStoreRepository{
		eventStore:       eventStore,
		eventConverter:   eventConverter,
		metadataProvider: metadataProvider,
	}
}

// Save a list of events to a given stream
func (r EventStoreRepository) Save(ctx context.Context, streamId store.StreamID, events event.List, version Version) error {
	expectedVersion := store.StreamVersion(version)
	var descriptors []store.EventDescriptor

	for _, e := range events {
		payload, err := r.eventConverter.ToEventPayload(e)
		if err != nil {
			return errors.Wrapf(err, "failed saving events to stream \"%s\"", streamId)
		}
		m := r.metadataProvider(e)

		descriptors = append(descriptors, store.EventDescriptor{
			ID:       store.EventID(uuid.New().String()),
			TypeName: e.TypeName(),
			Payload:  payload,
			Metadata: m,
		})
	}

	if err := r.eventStore.AppendToStream(ctx, streamId, descriptors, store.WithExpectedVersion(expectedVersion)); err != nil {
		return errors.Wrapf(err, "failed saving events to stream \"%s\"", streamId)
	}

	return nil
}

// Load a list of event from a given stream.
func (r EventStoreRepository) Load(ctx context.Context, id store.StreamID) (event.List, Version, error) {
	operationFailed := func(err error) error {
		return errors.Wrapf(err, "failed loading state from stream \"%s\"", id)
	}

	stream, err := r.eventStore.ReadFromStream(ctx, id, store.FromStart(), store.InForwardDirection())
	if err != nil {
		return nil, 0, operationFailed(err)
	}

	version := Version(store.InitialVersion)

	var events event.List
	for _, d := range stream.Descriptors {
		e, err := r.eventConverter.FromRecordedEventDescriptor(d)
		if err != nil {
			return nil, 0, operationFailed(err)
		}
		events = append(events, e)
		version = version.Incremented()
	}

	return nil, version, nil
}

// StateProjector is responsible for projecting an event onto a state.
type StateProjector[S any] func(s S, e event.Event) S

// StateHandler is responsible for handling a command for a given state and returning an event.List as a result
// of any doable state changes.
type StateHandler[S any, C command.Command] func(s S, c C) (event.List, error)

// IDProviderFromEvent is responsible for returning the ID of the stream associated with a given state for a given event.
type IDProviderFromEvent[E event.Event] func(e E) string

// EventStreamCreatingCommandHandler is an implementation of a command.Handler that automates
// the common boilerplate of command handlers that create new streams, by saving the resulting events to the event store.
func EventStreamCreatingCommandHandler[S any, C command.Command, E event.Event](
	repository EventStoreRepository,
	projector StateProjector[S],
	idProvider IDProviderFromEvent[E],
	handler StateHandler[S, C],
) command.HandlerFunc {
	if projector == nil {
		panic("projector cannot be nil")
	}

	if idProvider == nil {
		panic("idProvider cannot be nil")
	}

	if handler == nil {
		panic("handler cannot be nil")
	}

	return func(ctx context.Context, c command.Command) (any, error) {
		cmd := c.(C)
		var s S

		// Handle Command
		events, err := handler(s, cmd)
		if err != nil {
			return nil, err
		}

		// apply event on state
		var idProvided string
		for _, e := range events {
			s = projector(s, e)
			if _, ok := e.(E); ok {
				idProvided = idProvider(e.(E))
			}
		}

		// save to repository
		if err := repository.Save(ctx, store.StreamID(idProvided), events, Version(store.InitialVersion)); err != nil {
			return nil, err
		}

		// return list of events.
		return events, nil
	}
}

// IDProviderFromCommand is responsible for returning the ID of the stream associated with a given state for a given command.
type IDProviderFromCommand[C command.Command] func(c C) string

// EventStreamUpdatingCommandHandler is an implementation of a command.Handler that automates
// the common boilerplate of command handlers updating an event stream, by saving the resulting events to the event store.
func EventStreamUpdatingCommandHandler[S any, C command.Command](
	repository EventStoreRepository,
	projector StateProjector[S],
	idProvider IDProviderFromCommand[C],
	handler StateHandler[S, C],
) command.HandlerFunc {
	if projector == nil {
		panic("projector cannot be nil")
	}

	if idProvider == nil {
		panic("idProvider cannot be nil")
	}

	if handler == nil {
		panic("handler cannot be nil")
	}

	foldState := func(s S, events event.List) S {
		for _, e := range events {
			s = projector(s, e)
		}
		return s
	}

	return func(ctx context.Context, c command.Command) (any, error) {
		cmd := c.(C)
		idProvided := idProvider(cmd)
		var version Version
		var s S

		// If we have an ID, this means the command
		// is related to a certain state, otherwise, we are dealing with a creation command.
		if idProvided != "" {
			streamID := store.StreamID(idProvided)
			loaded, v, err := repository.Load(ctx, streamID)
			if err != nil {
				return nil, err
			}
			version = v

			// Load State from events.
			s = foldState(s, loaded)
		}

		// Handle Command
		events, err := handler(s, cmd)
		if err != nil {
			return nil, err
		}

		// apply event on state
		s = foldState(s, events)

		// save to repository
		if err := repository.Save(ctx, store.StreamID(idProvided), events, version); err != nil {
			return nil, err
		}

		// return list of events.
		return events, nil
	}
}
