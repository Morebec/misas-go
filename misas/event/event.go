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

package event

import (
	"github.com/morebec/misas-go/misas"
	"github.com/pkg/errors"
)

// PayloadTypeName Represents the unique type name of an Event for serialization and discriminatory purposes.
type PayloadTypeName string

// Event Represents a state change that has happened in the past.
// As such it is a common practice, and recommended adding a field indicating the date and time at which the event occurred.
type Event struct {
	Payload  Payload
	Metadata misas.Metadata
}

type Payload interface {
	TypeName() PayloadTypeName
}

func New(p Payload) Event {
	return NewWithMetadata(p, nil)
}

func NewWithMetadata(p Payload, m misas.Metadata) Event {
	return Event{Payload: p, Metadata: m}
}

// List Represents a list of events.
type List []Event

// NewList creates a new list of events.
func NewList(evts ...Event) []Event {
	return evts
}

// IsEmpty Indicates if the List is empty.
func (l List) IsEmpty() bool {
	return len(l) == 0
}

// Select returns a list of events according to a given predicate
func (l List) Select(p func(e Event) (bool, error)) (List, error) {
	var newEventList List
	for _, event := range l {
		keep, err := p(event)
		if err != nil {
			return nil, errors.Wrap(err, "failed selecting event in list")
		}

		if keep {
			newEventList = append(newEventList, event)
		}
	}

	return newEventList, nil
}

// SelectByTypeNames Returns a list of events with some event types filtered.
func (l List) SelectByTypeNames(tns ...PayloadTypeName) List {
	result, _ := l.Select(func(e Event) (bool, error) {
		for _, t := range tns {
			if t == e.Payload.TypeName() {
				return true, nil
			}
		}
		return false, nil
	})

	return result
}

// Exclude returns a list of events not containing some events according to a given predicate.
func (l List) Exclude(p func(e Event) (bool, error)) (List, error) {
	var newEventList List
	for _, event := range l {
		exclude, err := p(event)

		if err != nil {
			return nil, errors.Wrap(err, "failed excluding event from list")
		}

		if !exclude {
			newEventList = append(newEventList, event)
		}
	}

	return newEventList, nil
}

// ExcludeByTypeNames Returns a list of events with some event types excluded according to their type names.
func (l List) ExcludeByTypeNames(tns ...PayloadTypeName) List {
	result, _ := l.Exclude(func(e Event) (bool, error) {
		for _, t := range tns {
			if t == e.Payload.TypeName() {
				return true, nil
			}
		}
		return false, nil
	})

	return result
}

// First Returns the first event in the list
func (l List) First() *Event {
	if l.IsEmpty() {
		return nil
	}
	return &l[0]
}

// Last Returns the last event in the list
func (l List) Last() *Event {
	if l.IsEmpty() {
		return nil
	}
	return &l[len(l)-1]
}
