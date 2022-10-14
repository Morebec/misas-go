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

// Version represents the Version of an event sourced aggregate for concurrency checking.
type Version store.StreamVersion

// InitialVersion represents the initial version of an aggregate.
const InitialVersion = Version(store.InitialVersion)

// Incremented Returns the value of this version incremented by one.
func (v Version) Incremented() Version {
	return v + 1
}

// EventStoreRepository is a utility service responsible for storing and retrieving event.Event into a store.EventStore.
// it can be used as a base to implement repositories.
type EventStoreRepository struct {
	eventStore       store.EventStore
	eventConverter   *store.EventConverter
	streamPrefix     string
	metadataProvider func(e event.Event) misas.Metadata
}

// EventMetadataProvider represents a function responsible for providing metadata to events.
type EventMetadataProvider func(e event.Event) misas.Metadata

func NewEventStoreRepository(
	eventStore store.EventStore,
	eventConverter *store.EventConverter,
	streamPrefix string,
	metadataProvider EventMetadataProvider,
) EventStoreRepository {
	if metadataProvider == nil {
		metadataProvider = func(e event.Event) misas.Metadata {
			return misas.Metadata{}
		}
	}

	return EventStoreRepository{
		eventStore:       eventStore,
		eventConverter:   eventConverter,
		streamPrefix:     streamPrefix,
		metadataProvider: metadataProvider,
	}
}

// Save a list of events to a given stream
func (r EventStoreRepository) Save(ctx context.Context, streamID store.StreamID, events event.List, version Version) error {
	streamID = r.prefixStream(streamID)
	operationFailed := func(err error) error {
		return errors.Wrapf(err, "failed saving events to stream \"%s\"", streamID)
	}
	expectedVersion := store.StreamVersion(version)
	var descriptors []store.EventDescriptor

	for _, e := range events {
		payload, err := r.eventConverter.ToEventPayload(e)
		if err != nil {
			return operationFailed(err)
		}
		m := r.metadataProvider(e)

		descriptors = append(descriptors, store.EventDescriptor{
			ID:       store.EventID(uuid.New().String()),
			TypeName: e.TypeName(),
			Payload:  payload,
			Metadata: m,
		})
	}

	if err := r.eventStore.AppendToStream(ctx, streamID, descriptors, store.WithExpectedVersion(expectedVersion)); err != nil {
		return operationFailed(err)
	}

	return nil
}

// Load an event.List from a given stream with store.StreamID.
func (r EventStoreRepository) Load(ctx context.Context, streamID store.StreamID) (event.List, Version, error) {
	streamID = r.prefixStream(streamID)
	operationFailed := func(err error) error {
		return errors.Wrapf(err, "failed loading state from stream \"%s\"", streamID)
	}

	stream, err := r.eventStore.ReadFromStream(ctx, streamID, store.FromStart(), store.InForwardDirection())
	if err != nil {
		return nil, 0, operationFailed(err)
	}

	version := InitialVersion

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

// prefixes a stream with the stream prefix of this repository.
func (r EventStoreRepository) prefixStream(streamID store.StreamID) store.StreamID {
	return store.StreamID(r.streamPrefix + string(streamID))
}

// StateProjector is responsible for projecting a event.Event onto a state.
type StateProjector[S any] func(s S, e event.Event) S

// StateHandler is responsible for handling a command for a given state and returning an event.List as a result
// of any doable state changes.
type StateHandler[S any, C command.Command] func(s S, c C) (event.List, error)

// StreamIDProviderFromState is responsible for returning the ID of the stream associated with a given state.
type StreamIDProviderFromState[S any] func(s S) store.StreamID

// StreamIDProviderFromCommand is responsible for returning the ID of the stream associated with a given state
// for a given command.Command
type StreamIDProviderFromCommand[C command.Command] func(c C) store.StreamID

// ResponseProvider is responsible for returning a response after handling a command based on an event
type ResponseProvider[S any] func(s S) any

// EventStreamCreatingCommandHandler is a utility implementation of a command.Handler that automates
// the common boilerplate of command handlers that create new streams of events, by saving the resulting events
// to the event store.
func EventStreamCreatingCommandHandler[S any, C command.Command](
	repository EventStoreRepository,
	projector StateProjector[S],
	handler StateHandler[S, C],
	streamIdProvider StreamIDProviderFromState[S],
	responseProvider ResponseProvider[S],
) command.HandlerFunc {
	if projector == nil {
		panic("projector cannot be nil")
	}

	if streamIdProvider == nil {
		panic("streamIdProvider cannot be nil")
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

		// apply event to state
		for _, e := range events {
			s = projector(s, e)
		}

		// save events to repository
		streamID := streamIdProvider(s)
		if err := repository.Save(ctx, store.StreamID(streamID), events, InitialVersion); err != nil {
			return nil, err
		}

		return responseProvider(s), nil
	}
}

// EventStreamUpdatingCommandHandler is a utility implementation of a command.Handler that automates
// the common boilerplate of command handlers that update streams of events, by saving the resulting events
// to the event store.
func EventStreamUpdatingCommandHandler[S any, C command.Command](
	repository EventStoreRepository,
	projector StateProjector[S],
	handler StateHandler[S, C],
	streamIdProvider StreamIDProviderFromCommand[C],
	responseProvider ResponseProvider[S],
) command.HandlerFunc {
	if projector == nil {
		panic("projector cannot be nil")
	}

	if streamIdProvider == nil {
		panic("streamIdProvider cannot be nil")
	}

	if handler == nil {
		panic("handler cannot be nil")
	}

	if responseProvider == nil {
		responseProvider = func(S) any {
			return nil
		}
	}

	return func(ctx context.Context, c command.Command) (any, error) {
		cmd := c.(C)
		streamID := streamIdProvider(cmd)
		version := InitialVersion
		var s S

		// Load events
		loaded, v, err := repository.Load(ctx, streamID)
		if err != nil {
			return nil, err
		}
		version = v

		// Load State from events.
		for _, e := range loaded {
			s = projector(s, e)
		}

		// Handle Command
		events, err := handler(s, cmd)
		if err != nil {
			return nil, err
		}

		// apply event on state
		for _, e := range events {
			s = projector(s, e)
		}

		// save events to repository
		if err := repository.Save(ctx, streamID, events, version); err != nil {
			return nil, err
		}

		return responseProvider(s), nil
	}
}
