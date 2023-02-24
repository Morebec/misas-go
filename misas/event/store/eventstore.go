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
	"fmt"
	"github.com/google/uuid"
	"github.com/morebec/misas-go/misas"
	"github.com/morebec/misas-go/misas/event"
	"time"
)

type EventStore interface {

	// GlobalStreamID Returns the EventID of the stream representing the "all" stream for this event store.
	GlobalStreamID() StreamID

	// AppendToStream Appends events to a stream using a given set of options.
	// If the stream does not exist, it will findPayloadStruct implicitly created.
	// To enforce consistency boundaries when required, the AppendStreamOptions has the concept of an expected version,
	// where the current version of the stream is compared to this expected version. If they are not the same, this method will return
	// a ConcurrencyError
	AppendToStream(ctx context.Context, streamID StreamID, events []EventDescriptor, opts ...AppendToStreamOption) error

	// ReadFromStream Reads an event stream using a given set of options. If the stream does not exist, an error will be returned.
	ReadFromStream(ctx context.Context, streamID StreamID, opts ...ReadFromStreamOption) (StreamSlice, error)

	// TruncateStream Truncates a stream by removing some events in it using a given set of options.
	// To represent that fact, it should also append an event indicating this.
	// Depending on the underlying technology, this event can take many forms, and therefore this
	// interface does not prescribe any specific schema. It only prescribes that this truncation event be appended to the stream.
	// As a helper, the StreamTruncatedEvent can be used by implementors of the interface.
	// If the stream does not exist, will silently return.
	TruncateStream(ctx context.Context, streamID StreamID, opts ...TruncateStreamOption) error

	// DeleteStream permanently deletes a stream and its events, as if they had never happened in the system.
	// To represent that fact, it should also append an event indicating this in a technical stream.
	// Depending on the underlying technology, this event can take many forms, and therefore the event.Save
	// interface does not prescribe any specific schema. It only prescribes that this deletion event be appended to the stream.
	// As a helper, the StreamDeletedEvent can be used by implementors of the interface.
	// If the stream does not exist, will silently return.
	DeleteStream(ctx context.Context, id StreamID) error

	// SubscribeToStream Subscribes to a stream or returns an error, if the subscription could not be made.
	// If the stream does not exist, an error will be returned.
	SubscribeToStream(ctx context.Context, streamID StreamID, opts ...SubscribeToStreamOption) (Subscription, error)

	// StreamExists returns true if a stream exists, otherwise false.
	StreamExists(ctx context.Context, id StreamID) (bool, error)

	// GetStream Returns a given stream.
	// If the stream does not exist it is returned as an error.
	GetStream(ctx context.Context, id StreamID) (Stream, error)

	// Clear this event store
	Clear(ctx context.Context) error
}

// StreamID represents the EventID of a stream.
type StreamID string

// StreamVersion Represents the version of a stream.
type StreamVersion int64

const InitialVersion StreamVersion = -1

// SequenceNumber represents the number of an event in the global ordering of the event store.
type SequenceNumber int64

type Stream struct {
	ID             StreamID
	Version        StreamVersion
	InitialVersion StreamVersion
}

// EventID represents the unique identifier of an event in the store.
type EventID string

// NewEventID Generates a new unique EventID.
func NewEventID() EventID {
	return EventID(uuid.NewString())
}

// DescriptorPayload represents the payload of an event descriptor.
type DescriptorPayload map[string]any

// EventDescriptor Represents a wrapper around an event to be added to the store.
type EventDescriptor struct {
	ID       EventID
	TypeName event.PayloadTypeName
	Payload  DescriptorPayload
	Metadata misas.Metadata
}

// RecordedEventDescriptor represents an event descriptor for an event that was previously recorded in the store.
type RecordedEventDescriptor struct {
	ID             EventID
	TypeName       event.PayloadTypeName
	Payload        DescriptorPayload
	Metadata       misas.Metadata
	StreamID       StreamID
	Version        StreamVersion
	SequenceNumber SequenceNumber
	RecordedAt     time.Time
}

// StreamSlice Represents a slice of events that were read from a given stream.
type StreamSlice struct {
	StreamID    StreamID
	Descriptors []RecordedEventDescriptor
}

// First Returns the first descriptor in the slice.
func (s StreamSlice) First() RecordedEventDescriptor {
	if s.IsEmpty() {
		return RecordedEventDescriptor{}
	}
	return s.Descriptors[0]
}

// Last Returns the last descriptor in the slice.
func (s StreamSlice) Last() RecordedEventDescriptor {
	if s.IsEmpty() {
		return RecordedEventDescriptor{}
	}
	return s.Descriptors[len(s.Descriptors)-1]
}

// IsEmpty Indicates if the slice is empty or not
func (s StreamSlice) IsEmpty() bool {
	return len(s.Descriptors) == 0
}

// Length Returns the length of the slice
func (s StreamSlice) Length() int {
	return len(s.Descriptors)
}

// Select Returns a list of event descriptors according to a selection predicate.
func (s StreamSlice) Select(p func(event RecordedEventDescriptor) bool) []RecordedEventDescriptor {
	var events []RecordedEventDescriptor

	for _, d := range s.Descriptors {
		if p(d) {
			events = append(events, d)
		}
	}
	return events
}

// Reversed returns a copy of this slice with the events in reverse order.
func (s StreamSlice) Reversed() StreamSlice {
	for i, j := 0, len(s.Descriptors)-1; i < j; i, j = i+1, j-1 {
		s.Descriptors[i], s.Descriptors[j] = s.Descriptors[j], s.Descriptors[i]
	}

	return s
}

// StreamNotFoundError error representing the fact that a stream was not found.
type StreamNotFoundError struct {
	StreamID StreamID
}

func (s StreamNotFoundError) Error() string {
	return fmt.Sprintf("stream \"%s\" not found ", s.StreamID)
}

func NewStreamNotFoundError(id StreamID) error {
	return StreamNotFoundError{StreamID: id}
}

// IsStreamNotFoundError Indicates if a given error is a StreamNotFoundError or not.
func IsStreamNotFoundError(err error) bool {
	_, ok := err.(StreamNotFoundError)
	return ok
}

// IsConcurrencyError Indicates if a given error is a ConcurrencyError
func IsConcurrencyError(err error) bool {
	_, ok := err.(ConcurrencyError)
	return ok
}
