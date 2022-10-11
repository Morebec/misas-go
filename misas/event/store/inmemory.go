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
	"github.com/morebec/go-system/misas/clock"
	"github.com/pkg/errors"
)

type InMemoryEventStore struct {
	Clock             clock.Clock
	events            []RecordedEventDescriptor
	eventIds          map[EventID]struct{}
	streamVersionByID map[StreamID]StreamVersion
	subscriptions     []Subscription
}

func NewInMemoryEventStore(clock clock.Clock) *InMemoryEventStore {
	return &InMemoryEventStore{
		Clock:             clock,
		events:            []RecordedEventDescriptor{},
		eventIds:          map[EventID]struct{}{},
		streamVersionByID: map[StreamID]StreamVersion{},
		subscriptions:     []Subscription{},
	}
}

func (es *InMemoryEventStore) SubscribeToStream(ctx context.Context, streamID StreamID, opts ...SubscribeToStreamOption) (Subscription, error) {

	options := BuildSubscribeToStreamOptions(opts)
	errorChannel := make(chan error)
	eventChannel := make(chan RecordedEventDescriptor)
	closeChannel := make(chan bool, 1)
	subscription := *NewSubscription(eventChannel, errorChannel, closeChannel, streamID, options)
	es.subscriptions = append(es.subscriptions, subscription)

	go func() {
		var filterOptions []TypeNameFilterOption
		if options.EventTypeNameFilter != nil {
			if options.EventTypeNameFilter.Mode == Exclude {
				filterOptions = append(filterOptions, ExcludeEventTypeNames(options.EventTypeNameFilter.EventTypeNames...))
			} else {
				filterOptions = append(filterOptions, SelectEventTypeNames(options.EventTypeNameFilter.EventTypeNames...))
			}
		}
		// Read form position
		streamSlice, err := es.ReadFromStream(ctx, streamID, WithMaxCount(0), InForwardDirection(), WithReadingFilter(filterOptions...))
		if err != nil {
			return
		}

		// Send read events to the subscription
		for _, e := range streamSlice.Descriptors {
			eventChannel <- e
		}
	}()

	return subscription, nil
}

func (es *InMemoryEventStore) GlobalStreamID() StreamID {
	return "$all"
}

func (es *InMemoryEventStore) AppendToStream(ctx context.Context, streamID StreamID, descriptors []EventDescriptor, opts ...AppendToStreamOption) error {

	options := BuildAppendToStreamOptions(opts)
	if streamID == es.GlobalStreamID() {
		return errors.New("cannot append to virtual stream")
	}

	lastSeqNo := SequenceNumber(len(es.events) - 1)
	nextSeqNo := lastSeqNo

	streamVersion, found := es.streamVersionByID[streamID]
	if !found {
		streamVersion = InitialVersion
	}

	// Check concurrency
	if options.ExpectedVersion != nil {
		if streamVersion != *options.ExpectedVersion {
			return NewConcurrencyError(streamID, *options.ExpectedVersion, streamVersion)
		}
	}

	var recordedEvents []RecordedEventDescriptor
	for _, d := range descriptors {
		if _, found := es.eventIds[d.ID]; found {
			return errors.Errorf("duplicate event id encountered with \"%s\"", d.ID)
		}

		streamVersion++
		nextSeqNo++

		rd := RecordedEventDescriptor{
			ID:             d.ID,
			TypeName:       d.TypeName,
			Payload:        d.Payload,
			Metadata:       d.Metadata,
			StreamID:       streamID,
			Version:        streamVersion,
			RecordedAt:     es.Clock.Now(),
			SequenceNumber: nextSeqNo,
		}
		es.events = append(es.events, rd)
		recordedEvents = append(recordedEvents, rd)
	}

	es.streamVersionByID[streamID] = streamVersion

	// Notify subscribers
	go func() {
		for _, d := range recordedEvents {
			for _, sub := range es.subscriptions {
				if sub.streamID == es.GlobalStreamID() || sub.streamID == d.StreamID {
					sub.EmitEvent(d)
				}
			}
		}
	}()

	return nil
}

func (es *InMemoryEventStore) ReadFromStream(ctx context.Context, streamID StreamID, opts ...ReadFromStreamOption) (StreamSlice, error) {

	options := BuildReadFromStreamOptions(opts)
	isGlobalStream := streamID == es.GlobalStreamID()

	if !isGlobalStream {
		streamExists, err := es.StreamExists(ctx, streamID)
		if err != nil {
			return StreamSlice{}, err
		}

		if !streamExists {
			return StreamSlice{}, NewStreamNotFoundError(streamID)
		}
	}

	eventsOfStream := es.events
	if !isGlobalStream {
		eventsOfStream = StreamSlice{
			StreamID:    streamID,
			Descriptors: es.events,
		}.Select(func(descriptor RecordedEventDescriptor) bool {
			return streamID == descriptor.StreamID
		})
	}

	streamSlice := StreamSlice{
		StreamID:    streamID,
		Descriptors: eventsOfStream,
	}

	// Direction
	if options.Direction == Backward {
		streamSlice = streamSlice.Reversed()
	}

	// From position
	streamSlice = StreamSlice{
		StreamID: streamID,
		Descriptors: streamSlice.Select(func(descriptor RecordedEventDescriptor) bool {
			var eventPosition Position
			if streamID == es.GlobalStreamID() {
				eventPosition = Position(descriptor.SequenceNumber)
			} else {
				eventPosition = Position(descriptor.Version)
			}

			if options.Direction == Backward {
				return eventPosition < options.Position
			}

			return eventPosition > options.Position
		}),
	}

	// Type names
	if options.EventTypeNameFilter != nil {
		streamSlice = StreamSlice{
			StreamID: streamID,
			Descriptors: streamSlice.Select(func(descriptor RecordedEventDescriptor) bool {
				matchesFilter := false
				for _, tn := range options.EventTypeNameFilter.EventTypeNames {
					if tn == descriptor.TypeName {
						matchesFilter = true
						break
					}
				}

				if options.EventTypeNameFilter.Mode == Exclude {
					return !matchesFilter
				}

				return matchesFilter
			}),
		}
	}

	return streamSlice, nil
}

func (es *InMemoryEventStore) TruncateStream(ctx context.Context, streamID StreamID, opts ...TruncateStreamOption) error {
	options := BuildTruncateFromStreamOptions(opts)

	streamExists, err := es.StreamExists(ctx, streamID)
	if err != nil {
		return err
	}

	if !streamExists {
		return NewStreamNotFoundError(streamID)
	}

	es.events = StreamSlice{
		StreamID:    es.GlobalStreamID(),
		Descriptors: es.events,
	}.Select(func(descriptor RecordedEventDescriptor) bool {
		if descriptor.StreamID != streamID {
			return true
		}
		return Position(descriptor.Version) >= options.BeforePosition
	})

	return nil
}

func (es *InMemoryEventStore) DeleteStream(ctx context.Context, id StreamID) error {
	exists, err := es.StreamExists(ctx, id)
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	es.events = StreamSlice{
		StreamID:    es.GlobalStreamID(),
		Descriptors: es.events,
	}.Select(func(descriptor RecordedEventDescriptor) bool {
		return descriptor.StreamID != id
	})

	delete(es.streamVersionByID, id)

	return nil
}

func (es *InMemoryEventStore) Clear(ctx context.Context) error {
	es.events = []RecordedEventDescriptor{}
	es.eventIds = map[EventID]struct{}{}
	es.streamVersionByID = map[StreamID]StreamVersion{}

	return nil
}

func (es *InMemoryEventStore) StreamExists(ctx context.Context, id StreamID) (bool, error) {
	_, found := es.streamVersionByID[id]
	return found, nil
}

func (es *InMemoryEventStore) GetStream(ctx context.Context, id StreamID) (Stream, error) {

	min := StreamVersion(Start)
	max := min

	for _, d := range es.events {
		if d.StreamID == id {
			if d.Version > max {
				max = d.Version
			}

			if d.Version < min {
				min = d.Version
			}
		}
	}

	if min == max {
		return Stream{}, NewStreamNotFoundError(id)
	} else if min == StreamVersion(Start) {
		min = 0
	}

	return Stream{
		ID:             id,
		Version:        min,
		InitialVersion: max,
	}, nil
}
