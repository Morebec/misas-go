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

type EventSourcedAggregate interface {
	Apply(e event.Event)
}

// Version represents the Version of an event sourced aggregate for concurrency checking.
type Version store.StreamVersion

// InitialVersion represents the initial version of an aggregate at creation time.
const InitialVersion = Version(store.InitialVersion)

// Incremented Returns the value of a version incremented by one.
func (v Version) Incremented() Version {
	return v + 1
}

// EventStoreRepository is a utility service responsible for storing and retrieving event.Event into a store.EventStore.
// it can be used as a base to implement repositories for specific aggregates types.
type EventStoreRepository struct {
	eventStore     store.EventStore
	eventConverter *store.EventConverter
	// Prefix to add to the event stream ID.
	streamPrefix     string
	metadataProvider func(e event.Event) misas.Metadata
}

// EventMetadataProvider represents a function responsible for providing metadata to events.
type EventMetadataProvider func(e event.Event) misas.Metadata

// NewEventStoreRepository allows creating an EventStoreRepository.
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

// Save a list of events to a given stream. If the EventStoreRepository uses a prefix, it will be appended to the value
// provided as an argument to this function call.
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

// Load an event.List from a given stream with store.StreamID. If the EventStoreRepository uses a prefix, it will be appended to the value
// provided as an argument to this function call.
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
