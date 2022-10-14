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

package esdb

import (
	"context"
	"github.com/google/uuid"
	"github.com/morebec/misas-go/misas"
	"github.com/morebec/misas-go/misas/event"
	"github.com/morebec/misas-go/misas/event/store"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type esdbSQLUnitTestPassedEvent struct {
	TestName string
}

func (u esdbSQLUnitTestPassedEvent) TypeName() event.TypeName {
	return "unit_test.passed"
}

func buildEventStore() *EventStore {
	ctx := context.Background()

	s, err := NewEventStoreFromConnectionString("esdb+discover://localhost:2113?keepAliveTimeout=10000&keepAliveInterval=10000&tls=false")
	if err != nil {
		panic(err)
	}

	if err := s.Open(ctx); err != nil {
		panic(err)
	}

	if err := s.Clear(ctx); err != nil {
		panic(err)
	}

	return s
}

func TestEventStore_OpenConnection(t *testing.T) {
	assert.NotPanics(t, func() {
		_ = buildEventStore()
	})
}

func TestEventStore_CloseConnection(t *testing.T) {
	st := buildEventStore()
	err := st.Close(context.Background())
	assert.NoError(t, err)
}

func TestEventStore_AppendToStream(t *testing.T) {
	st := buildEventStore()

	streamID := store.StreamID("unit_test")
	eventId1 := store.EventID(uuid.NewString())
	eventId2 := store.EventID(uuid.NewString())
	eventId3 := store.EventID(uuid.NewString())

	err := st.AppendToStream(context.Background(), streamID, []store.EventDescriptor{
		{
			ID:       eventId1,
			TypeName: esdbSQLUnitTestPassedEvent{}.TypeName(),
			Payload:  store.EventPayload{"TestName": "AppendToStream"},
			Metadata: misas.Metadata{"hello": "world"},
		},
		{
			ID:       eventId2,
			TypeName: esdbSQLUnitTestPassedEvent{}.TypeName(),
			Payload:  store.EventPayload{"TestName": "AppendToStream"},
			Metadata: misas.Metadata{},
		},
		{
			ID:       eventId3,
			TypeName: esdbSQLUnitTestPassedEvent{}.TypeName(),
			Payload:  store.EventPayload{"TestName": "AppendToStream"},
			Metadata: misas.Metadata{},
		},
	}, store.WithOptimisticConcurrencyCheckDisabled())
	assert.NoError(t, err)

	// Read forward
	events, err := st.ReadFromStream(context.Background(), streamID, store.FromStart(), store.InForwardDirection())
	assert.NoError(t, err)
	assert.Len(t, events.Descriptors, 3)
	assert.Equal(t, esdbSQLUnitTestPassedEvent{}.TypeName(), events.First().TypeName)
	assert.Equal(t, eventId1, events.First().ID)
	assert.Equal(t, esdbSQLUnitTestPassedEvent{}.TypeName(), events.Last().TypeName)
	assert.Equal(t, eventId3, events.Last().ID)
	assert.Equal(t, misas.Metadata{"hello": "world"}, events.First().Metadata)
}

func TestEventStore_ReadFromStream(t *testing.T) {
	st := buildEventStore()

	streamID := store.StreamID("unit_test")
	err := st.AppendToStream(context.Background(), streamID, []store.EventDescriptor{
		{
			ID:       "event#1",
			TypeName: esdbSQLUnitTestPassedEvent{}.TypeName(),
			Payload:  store.EventPayload{"TestName": "ReadFromStream"},
			Metadata: misas.Metadata{},
		},
		{
			ID:       "event#2",
			TypeName: esdbSQLUnitTestPassedEvent{}.TypeName(),
			Payload:  store.EventPayload{"TestName": "ReadFromStream"},
			Metadata: misas.Metadata{},
		},
		{
			ID:       "event#3",
			TypeName: esdbSQLUnitTestPassedEvent{}.TypeName(),
			Payload:  store.EventPayload{"TestName": "ReadFromStream"},
			Metadata: misas.Metadata{},
		},
	}, store.WithOptimisticConcurrencyCheckDisabled())
	assert.Nil(t, err)

	// Read forward stream from start
	events, err := st.ReadFromStream(context.Background(), streamID, store.FromStart(), store.InForwardDirection())
	assert.Nil(t, err)
	assert.Len(t, events.Descriptors, 3)
	assert.Equal(t, events.First().TypeName, esdbSQLUnitTestPassedEvent{}.TypeName())
	assert.Equal(t, events.First().ID, store.EventID("event#1"))
	assert.Equal(t, events.Last().TypeName, esdbSQLUnitTestPassedEvent{}.TypeName())
	assert.Equal(t, events.Last().ID, store.EventID("event#3"))

	// Read forward global from start
	events, err = st.ReadFromStream(context.Background(), st.GlobalStreamID(), store.FromStart(), store.InForwardDirection())
	assert.Nil(t, err)
	assert.Len(t, events.Descriptors, 3)
	assert.Equal(t, events.First().TypeName, esdbSQLUnitTestPassedEvent{}.TypeName())
	assert.Equal(t, events.First().ID, store.EventID("event#1"))
	assert.Equal(t, events.Last().TypeName, esdbSQLUnitTestPassedEvent{}.TypeName())
	assert.Equal(t, events.Last().ID, store.EventID("event#3"))

	// Read forward stream from position
	events, err = st.ReadFromStream(context.Background(), streamID, store.From(0), store.InForwardDirection())
	assert.Nil(t, err)
	assert.Len(t, events.Descriptors, 2)
	assert.Equal(t, events.First().TypeName, esdbSQLUnitTestPassedEvent{}.TypeName())
	assert.Equal(t, events.First().ID, store.EventID("event#2"))
	assert.Equal(t, events.Last().TypeName, esdbSQLUnitTestPassedEvent{}.TypeName())
	assert.Equal(t, events.Last().ID, store.EventID("event#3"))

	// Read backwards from stream
	events, err = st.ReadFromStream(context.Background(), streamID, store.FromEnd(), store.InBackwardDirection())
	assert.Nil(t, err)
	assert.Len(t, events.Descriptors, 3)
	assert.Equal(t, events.First().TypeName, esdbSQLUnitTestPassedEvent{}.TypeName())
	assert.Equal(t, events.First().ID, store.EventID("event#3"))
	assert.Equal(t, events.Last().TypeName, esdbSQLUnitTestPassedEvent{}.TypeName())
	assert.Equal(t, events.Last().ID, store.EventID("event#1"))

	// Read backwards global
	events, err = st.ReadFromStream(context.Background(), st.GlobalStreamID(), store.FromEnd(), store.InBackwardDirection())
	assert.Nil(t, err)
	assert.Len(t, events.Descriptors, 3)
	assert.Equal(t, events.First().TypeName, esdbSQLUnitTestPassedEvent{}.TypeName())
	assert.Equal(t, events.First().ID, store.EventID("event#3"))
	assert.Equal(t, events.Last().TypeName, esdbSQLUnitTestPassedEvent{}.TypeName())
	assert.Equal(t, events.Last().ID, store.EventID("event#1"))

	// TODO Test event type name filter.
}

func TestEventStore_Clear(t *testing.T) {
	st := buildEventStore()

	streamID := store.StreamID("unit_test")
	err := st.AppendToStream(context.Background(), streamID, []store.EventDescriptor{
		{
			ID:       store.EventID(uuid.New().String()),
			TypeName: esdbSQLUnitTestPassedEvent{}.TypeName(),
			Payload:  store.EventPayload{"TestName": "Clear"},
			Metadata: misas.Metadata{},
		},
	})
	assert.NoError(t, err)

	err = st.Clear(context.Background())
	assert.NoError(t, err)

	_, err = st.ReadFromStream(context.Background(), streamID, store.FromStart(), store.InForwardDirection())
	assert.Error(t, err)
}

func TestEventStore_DeleteStream(t *testing.T) {
	st := buildEventStore()

	streamID := store.StreamID("unit_test")
	err := st.AppendToStream(context.Background(), streamID, []store.EventDescriptor{
		{
			ID:       store.EventID(uuid.New().String()),
			TypeName: esdbSQLUnitTestPassedEvent{}.TypeName(),
			Payload:  store.EventPayload{},
			Metadata: misas.Metadata{},
		},
	})
	assert.NoError(t, err)

	err = st.DeleteStream(context.Background(), streamID)
	assert.NoError(t, err)

	_, err = st.ReadFromStream(context.Background(), streamID, store.FromStart(), store.InForwardDirection())
	assert.Error(t, err)
}

func TestEventStore_GetStream(t *testing.T) {
	st := buildEventStore()

	streamID := store.StreamID("unit_test")

	_, err := st.GetStream(context.Background(), streamID)
	assert.Error(t, err)

	err = st.AppendToStream(context.Background(), streamID, []store.EventDescriptor{
		{
			ID:       store.EventID(uuid.New().String()),
			TypeName: esdbSQLUnitTestPassedEvent{}.TypeName(),
			Payload:  store.EventPayload{"TestName": "GetStream"},
			Metadata: misas.Metadata{},
		},
	})
	assert.NoError(t, err)

	stream, err := st.GetStream(context.Background(), streamID)
	assert.NoError(t, err)
	assert.Equal(t, store.Stream{
		ID:             streamID,
		Version:        0,
		InitialVersion: 0,
	}, stream)
}

func TestEventStore_StreamExists(t *testing.T) {
	st := buildEventStore()

	streamID := store.StreamID("unit_test")

	exists, err := st.StreamExists(context.Background(), streamID)
	assert.NoError(t, err)
	assert.False(t, exists)

	err = st.AppendToStream(context.Background(), streamID, []store.EventDescriptor{
		{
			ID:       store.EventID(uuid.New().String()),
			TypeName: esdbSQLUnitTestPassedEvent{}.TypeName(),
			Payload:  store.EventPayload{"TestName": "StreamExists"},
			Metadata: misas.Metadata{},
		},
	})
	assert.NoError(t, err)

	exists, err = st.StreamExists(context.Background(), streamID)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestEventStore_SubscribeToStream(t *testing.T) {
	st := buildEventStore()
	streamID := store.StreamID("unit_test")
	err := st.AppendToStream(context.Background(), streamID, []store.EventDescriptor{
		{
			ID:       "event#1",
			TypeName: esdbSQLUnitTestPassedEvent{}.TypeName(),
			Payload:  store.EventPayload{"TestName": "SubscribeToStream"},
			Metadata: misas.Metadata{},
		},
		{
			ID:       "event#2",
			TypeName: esdbSQLUnitTestPassedEvent{}.TypeName(),
			Payload:  store.EventPayload{"TestName": "SubscribeToStream"},
			Metadata: misas.Metadata{},
		},
		{
			ID:       "event#3",
			TypeName: esdbSQLUnitTestPassedEvent{}.TypeName(),
			Payload:  store.EventPayload{"TestName": "SubscribeToStream"},
			Metadata: misas.Metadata{},
		},
	})
	assert.NoError(t, err)

	subscription, err := st.SubscribeToStream(context.Background(), streamID)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)

	// New event
	err = st.AppendToStream(context.Background(), streamID, []store.EventDescriptor{
		{
			ID:       "event#4",
			TypeName: esdbSQLUnitTestPassedEvent{}.TypeName(),
			Payload:  store.EventPayload{"TestName": "SubscribeToStream"},
			Metadata: misas.Metadata{},
		},
	})
	assert.NoError(t, err)

	e := <-subscription.EventChannel()
	assert.Equal(t, store.EventID("event#4"), e.ID)

	err = subscription.Close()
	assert.NoError(t, err)
}

func TestEventStore_TruncateStream(t *testing.T) {
	st := buildEventStore()

	streamID := store.StreamID("unit_test")
	err := st.AppendToStream(context.Background(), streamID, []store.EventDescriptor{
		{
			ID:       "event#1",
			TypeName: esdbSQLUnitTestPassedEvent{}.TypeName(),
			Payload:  store.EventPayload{"TestName": "TruncateStream"},
			Metadata: misas.Metadata{},
		},
		{
			ID:       "event#2",
			TypeName: esdbSQLUnitTestPassedEvent{}.TypeName(),
			Payload:  store.EventPayload{"TestName": "TruncateStream"},
			Metadata: misas.Metadata{},
		},
		{
			ID:       "event#3",
			TypeName: esdbSQLUnitTestPassedEvent{}.TypeName(),
			Payload:  store.EventPayload{"TestName": "TruncateStream"},
			Metadata: misas.Metadata{},
		},
	})
	assert.NoError(t, err)

	err = st.TruncateStream(context.Background(), streamID, store.BeforePosition(1))
	assert.NoError(t, err)

	events, err := st.ReadFromStream(context.Background(), streamID, store.FromStart(), store.InForwardDirection())
	assert.Len(t, events.Descriptors, 2)
}
