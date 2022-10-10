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

package store

import (
	"context"
	"github.com/google/uuid"
	"github.com/jwillp/go-system/misas"
	"github.com/jwillp/go-system/misas/clock"
	"github.com/jwillp/go-system/misas/event"
	"github.com/stretchr/testify/assert"
	"testing"
)

const InMemoryUnitTestPassedEventTypeName event.TypeName = "unit_test.passed"

func TestInMemoryEventStore_AppendToStream(t *testing.T) {
	store := NewInMemoryEventStore(clock.UTCClock{})

	streamID := StreamID("unit_test")
	err := store.AppendToStream(context.Background(), streamID, []EventDescriptor{
		{
			ID:       "event#1",
			TypeName: InMemoryUnitTestPassedEventTypeName,
			Payload:  EventPayload{},
			Metadata: misas.Metadata{"hello": "world"},
		},
		{
			ID:       "event#2",
			TypeName: InMemoryUnitTestPassedEventTypeName,
			Payload:  EventPayload{},
			Metadata: misas.Metadata{},
		},
		{
			ID:       "event#3",
			TypeName: InMemoryUnitTestPassedEventTypeName,
			Payload:  EventPayload{},
			Metadata: misas.Metadata{},
		},
	})
	assert.NoError(t, err)

	// Read forward
	events, err := store.ReadFromStream(context.Background(), streamID, FromStart(), InForwardDirection())
	assert.NoError(t, err)
	assert.Len(t, events.Descriptors, 3)
	assert.Equal(t, events.First().TypeName, InMemoryUnitTestPassedEventTypeName)
	assert.Equal(t, events.First().ID, EventID("event#1"))
	assert.Equal(t, events.Last().TypeName, InMemoryUnitTestPassedEventTypeName)
	assert.Equal(t, events.Last().ID, EventID("event#3"))
	assert.Equal(t, misas.Metadata{"hello": "world"}, events.First().Metadata)
}

func TestInMemoryEventStore_ReadFromStream(t *testing.T) {
	store := NewInMemoryEventStore(clock.UTCClock{})

	streamID := StreamID("unit_test")
	err := store.AppendToStream(context.Background(), streamID, []EventDescriptor{
		{
			ID:       "event#1",
			TypeName: InMemoryUnitTestPassedEventTypeName,
			Payload:  EventPayload{},
			Metadata: misas.Metadata{},
		},
		{
			ID:       "event#2",
			TypeName: InMemoryUnitTestPassedEventTypeName,
			Payload:  EventPayload{},
			Metadata: misas.Metadata{},
		},
		{
			ID:       "event#3",
			TypeName: InMemoryUnitTestPassedEventTypeName,
			Payload:  EventPayload{},
			Metadata: misas.Metadata{},
		},
	})
	assert.Nil(t, err)

	// Read forward stream from start
	events, err := store.ReadFromStream(context.Background(), streamID, FromStart(), InForwardDirection())
	assert.Nil(t, err)
	assert.Len(t, events.Descriptors, 3)
	assert.Equal(t, events.First().TypeName, InMemoryUnitTestPassedEventTypeName)
	assert.Equal(t, events.First().ID, EventID("event#1"))
	assert.Equal(t, events.Last().TypeName, InMemoryUnitTestPassedEventTypeName)
	assert.Equal(t, events.Last().ID, EventID("event#3"))

	// Read forward global from start
	events, err = store.ReadFromStream(context.Background(), store.GlobalStreamID(), FromStart(), InForwardDirection())
	assert.Nil(t, err)
	assert.Len(t, events.Descriptors, 3)
	assert.Equal(t, events.First().TypeName, InMemoryUnitTestPassedEventTypeName)
	assert.Equal(t, events.First().ID, EventID("event#1"))
	assert.Equal(t, events.Last().TypeName, InMemoryUnitTestPassedEventTypeName)
	assert.Equal(t, events.Last().ID, EventID("event#3"))

	// Read forward stream from position
	events, err = store.ReadFromStream(context.Background(), streamID, From(0), InForwardDirection())
	assert.Nil(t, err)
	assert.Len(t, events.Descriptors, 2)
	assert.Equal(t, events.First().TypeName, InMemoryUnitTestPassedEventTypeName)
	assert.Equal(t, events.First().ID, EventID("event#2"))
	assert.Equal(t, events.Last().TypeName, InMemoryUnitTestPassedEventTypeName)
	assert.Equal(t, events.Last().ID, EventID("event#3"))

	// Read backwards from stream
	events, err = store.ReadFromStream(context.Background(), streamID, FromEnd(), InBackwardDirection())
	assert.Nil(t, err)
	assert.Len(t, events.Descriptors, 3)
	assert.Equal(t, events.First().TypeName, InMemoryUnitTestPassedEventTypeName)
	assert.Equal(t, events.First().ID, EventID("event#3"))
	assert.Equal(t, events.Last().TypeName, InMemoryUnitTestPassedEventTypeName)
	assert.Equal(t, events.Last().ID, EventID("event#1"))

	// Read backwards global
	events, err = store.ReadFromStream(context.Background(), store.GlobalStreamID(), FromEnd(), InBackwardDirection())
	assert.Nil(t, err)
	assert.Len(t, events.Descriptors, 3)
	assert.Equal(t, events.First().TypeName, InMemoryUnitTestPassedEventTypeName)
	assert.Equal(t, events.First().ID, EventID("event#3"))
	assert.Equal(t, events.Last().TypeName, InMemoryUnitTestPassedEventTypeName)
	assert.Equal(t, events.Last().ID, EventID("event#1"))
}

func TestInMemoryEventStore_Clear(t *testing.T) {
	store := NewInMemoryEventStore(clock.UTCClock{})

	streamID := StreamID("unit_test")
	err := store.AppendToStream(context.Background(), streamID, []EventDescriptor{
		{
			ID:       EventID(uuid.New().String()),
			TypeName: InMemoryUnitTestPassedEventTypeName,
			Payload:  EventPayload{},
			Metadata: misas.Metadata{},
		},
	})
	assert.NoError(t, err)

	err = store.Clear(context.Background())
	assert.NoError(t, err)

	_, err = store.ReadFromStream(context.Background(), streamID, FromStart(), InForwardDirection())
	assert.Error(t, err)
}

func TestInMemoryEventStore_DeleteStream(t *testing.T) {
	store := NewInMemoryEventStore(clock.UTCClock{})

	streamID := StreamID("unit_test")
	err := store.AppendToStream(context.Background(), streamID, []EventDescriptor{
		{
			ID:       EventID(uuid.New().String()),
			TypeName: InMemoryUnitTestPassedEventTypeName,
			Payload:  EventPayload{},
			Metadata: misas.Metadata{},
		},
	})
	assert.NoError(t, err)

	err = store.DeleteStream(context.Background(), streamID)
	assert.NoError(t, err)

	_, err = store.ReadFromStream(context.Background(), streamID, FromStart(), InForwardDirection())
	assert.Error(t, err)
}

func TestInMemoryEventStore_GetStream(t *testing.T) {
	store := NewInMemoryEventStore(clock.UTCClock{})

	streamID := StreamID("unit_test")

	_, err := store.GetStream(context.Background(), streamID)
	assert.Error(t, err)

	err = store.AppendToStream(context.Background(), streamID, []EventDescriptor{
		{
			ID:       EventID(uuid.New().String()),
			TypeName: InMemoryUnitTestPassedEventTypeName,
			Payload:  EventPayload{},
			Metadata: misas.Metadata{},
		},
	})
	assert.NoError(t, err)

	stream, err := store.GetStream(context.Background(), streamID)
	assert.NoError(t, err)
	assert.Equal(t, Stream{
		ID:             streamID,
		Version:        0,
		InitialVersion: 0,
	}, stream)
}

func TestInMemoryEventStore_StreamExists(t *testing.T) {
	store := NewInMemoryEventStore(clock.UTCClock{})

	streamID := StreamID("unit_test")

	exists, err := store.StreamExists(context.Background(), streamID)
	assert.NoError(t, err)
	assert.False(t, exists)

	err = store.AppendToStream(context.Background(), streamID, []EventDescriptor{
		{
			ID:       EventID(uuid.New().String()),
			TypeName: InMemoryUnitTestPassedEventTypeName,
			Payload:  EventPayload{},
			Metadata: misas.Metadata{},
		},
	})
	assert.NoError(t, err)

	exists, err = store.StreamExists(context.Background(), streamID)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestInMemoryEventStore_SubscribeToStreams(t *testing.T) {
	store := NewInMemoryEventStore(clock.UTCClock{})
	streamID := StreamID("unit_test")
	err := store.AppendToStream(context.Background(), streamID, []EventDescriptor{
		{
			ID:       "event#1",
			TypeName: InMemoryUnitTestPassedEventTypeName,
			Payload:  EventPayload{},
			Metadata: misas.Metadata{},
		},
		{
			ID:       "event#2",
			TypeName: InMemoryUnitTestPassedEventTypeName,
			Payload:  EventPayload{},
			Metadata: misas.Metadata{},
		},
		{
			ID:       "event#3",
			TypeName: InMemoryUnitTestPassedEventTypeName,
			Payload:  EventPayload{},
			Metadata: misas.Metadata{},
		},
	})
	assert.NoError(t, err)

	subscription, err := store.SubscribeToStream(context.Background(), streamID)
	assert.NoError(t, err)

	// Catch up
	e := <-subscription.EventChannel()
	assert.Equal(t, EventID("event#2"), e.ID)

	e = <-subscription.EventChannel()
	assert.Equal(t, EventID("event#3"), e.ID)

	// New event
	err = store.AppendToStream(context.Background(), streamID, []EventDescriptor{
		{
			ID:       "event#4",
			TypeName: InMemoryUnitTestPassedEventTypeName,
			Payload:  EventPayload{},
			Metadata: misas.Metadata{},
		},
	})
	assert.NoError(t, err)

	e = <-subscription.EventChannel()
	assert.Equal(t, EventID("event#4"), e.ID)
}

func TestInMemoryEventStore_TruncateStream(t *testing.T) {
	store := NewInMemoryEventStore(clock.UTCClock{})

	streamID := StreamID("unit_test")
	err := store.AppendToStream(context.Background(), streamID, []EventDescriptor{
		{
			ID:       "event#1",
			TypeName: InMemoryUnitTestPassedEventTypeName,
			Payload:  EventPayload{},
			Metadata: misas.Metadata{},
		},
		{
			ID:       "event#2",
			TypeName: InMemoryUnitTestPassedEventTypeName,
			Payload:  EventPayload{},
			Metadata: misas.Metadata{},
		},
		{
			ID:       "event#3",
			TypeName: InMemoryUnitTestPassedEventTypeName,
			Payload:  EventPayload{},
			Metadata: misas.Metadata{},
		},
	})
	assert.NoError(t, err)

	err = store.TruncateStream(context.Background(), streamID, BeforePosition(1))
	assert.NoError(t, err)

	events, err := store.ReadFromStream(context.Background(), streamID, FromStart(), InForwardDirection())
	assert.Len(t, events.Descriptors, 2)
}
