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
	"github.com/morebec/misas-go/misas/event"
	"math"
)

// Position (See ReadFromStreamOptions) Represents a position in the event store depending on where it is used.
// It represents either a sequence number for the global stream or a version number for all the other streams.
// The event corresponding to this position will not be included in the StreamSlice.
type Position int

const (
	Start = Position(InitialVersion)
	End   = Position(math.MaxInt)
)

// Direction in which the reading should be performed.
type Direction string

const (
	Forward  Direction = "FORWARD"
	Backward Direction = "BACKWARD"
)

type TypeNameFilterMode string

const (
	Select  TypeNameFilterMode = "SELECT"
	Exclude TypeNameFilterMode = "EXCLUDE"
)

// TypeNameFilter Represents a filter on the event type nam
type TypeNameFilter struct {
	Mode           TypeNameFilterMode
	EventTypeNames []event.TypeName
}

// ReadFromStreamOptions UpcastableEventPayload structure representing the options that can be used to read from a stream.
type ReadFromStreamOptions struct {
	Position            Position
	MaxCount            int
	Direction           Direction
	EventTypeNameFilter *TypeNameFilter
}

type ReadFromStreamOption func(ro *ReadFromStreamOptions)

// FromStart Allows specifying the read operation should start at the start of the stream.
func FromStart() ReadFromStreamOption {
	return From(Start)
}

// FromEnd Allows specifying the read operation should start at the end of the stream.
func FromEnd() ReadFromStreamOption {
	return From(End)
}

// From Allows specifying the read operation should be performed from a given position. Meaning that all events following
// the given position will be returned.
func From(p Position) ReadFromStreamOption {
	return func(ro *ReadFromStreamOptions) {
		ro.Position = p
	}
}

// InForwardDirection Allows specifying the read operation should be performed from oldest to latest.
func InForwardDirection() ReadFromStreamOption {
	return InDirection(Forward)
}

// InBackwardDirection  Allows specifying the read operation should be performed from latest to oldest.
func InBackwardDirection() ReadFromStreamOption {
	return InDirection(Backward)
}

func InDirection(d Direction) ReadFromStreamOption {
	return func(ro *ReadFromStreamOptions) {
		ro.Direction = d
	}
}

// WithMaxCount specifies the maximum number of events to return.
func WithMaxCount(maxCount int) ReadFromStreamOption {
	return func(ro *ReadFromStreamOptions) {
		ro.MaxCount = maxCount
	}
}
func LastEvent() ReadFromStreamOption {
	return func(ro *ReadFromStreamOptions) {
		ro.Direction = Backward
		ro.Position = End
		ro.MaxCount = 1
	}
}

type TypeNameFilterOption func(filter *TypeNameFilter)

func WithReadingFilter(opts ...TypeNameFilterOption) ReadFromStreamOption {
	return func(ro *ReadFromStreamOptions) {
		if len(opts) == 0 {
			ro.EventTypeNameFilter = nil
		} else {
			ro.EventTypeNameFilter = &TypeNameFilter{
				Mode:           Select,
				EventTypeNames: nil,
			}
			for _, opt := range opts {
				opt(ro.EventTypeNameFilter)
			}
		}
	}
}

// ExcludeEventTypeNames Allows specifying that only events of some given types should not be read.
func ExcludeEventTypeNames(typeNames ...event.TypeName) TypeNameFilterOption {
	return func(filter *TypeNameFilter) {
		filter.EventTypeNames = typeNames
		filter.Mode = Exclude
	}
}

// SelectEventTypeNames Allows specifying that only events that are of some given types should be read.
func SelectEventTypeNames(typeNames ...event.TypeName) TypeNameFilterOption {
	return func(filter *TypeNameFilter) {
		filter.EventTypeNames = typeNames
		filter.Mode = Select
	}
}

func BuildReadFromStreamOptions(opts []ReadFromStreamOption) *ReadFromStreamOptions {
	options := &ReadFromStreamOptions{
		Position:            0,
		MaxCount:            0,
		Direction:           Forward,
		EventTypeNameFilter: nil,
	}
	for _, opt := range opts {
		opt(options)
	}
	return options
}
