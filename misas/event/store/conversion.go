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
	"encoding/json"
	"github.com/morebec/misas-go/misas/event"
	"github.com/pkg/errors"
	"reflect"
)

// EventConverter is a service responsible for converting event.Event to EventDescriptor and RecordedEventDescriptor back to event.Event.
// Internally it relies on mapping the empty value of an event.Event to its event.PayloadTypeName so that it can read the event.PayloadTypeName
// of a given RecordedEventDescriptor to have the right in memory representation (struct) of the event.Event.
type EventConverter struct {
	events map[event.PayloadTypeName]reflect.Type
}

func NewEventConverter() *EventConverter {
	ec := &EventConverter{map[event.PayloadTypeName]reflect.Type{}}
	ec.RegisterEvent(StreamTruncatedEvent{})
	return ec
}

// ToEventPayload converts an event.Event to an DescriptorPayload to be used with an EventDescriptor.
func (c *EventConverter) ToEventPayload(evt event.Event) (DescriptorPayload, error) {
	marshal, err := json.Marshal(evt.Payload)
	if err != nil {
		return nil, errors.Wrapf(err, "failed converting event \"%s\" to DescriptorPayload", evt.Payload.TypeName())
	}

	var payload DescriptorPayload
	if err := json.Unmarshal(marshal, &payload); err != nil {
		return nil, errors.Wrapf(err, "failed converting event \"%s\" to DescriptorPayload", evt.Payload.TypeName())
	}

	return payload, nil
}

// FromRecordedEventDescriptor loads an event.Event from a RecordedEventDescriptor
func (c *EventConverter) FromRecordedEventDescriptor(descriptor RecordedEventDescriptor) (event.Event, error) {
	evt, err := c.find(descriptor.TypeName)
	if err != nil {
		return event.Event{}, errors.Wrapf(err, "failed loading event \"%s/%s\" in memory", descriptor.TypeName, descriptor.ID)
	}

	marshal, err := json.Marshal(descriptor.Payload)
	if err != nil {
		return event.Event{}, errors.Wrapf(err, "failed loading event \"%s/%s\" in memory", descriptor.TypeName, descriptor.ID)
	}

	if err := json.Unmarshal(marshal, evt); err != nil {
		return event.Event{}, errors.Wrapf(err, "failed loading event \"%s/%s\" in memory", descriptor.TypeName, descriptor.ID)
	}

	// Convert *event.Event to event.Event
	payload := reflect.ValueOf(evt).Elem().Interface().(event.Payload)

	metadata := descriptor.Metadata
	metadata.Set("id", string(descriptor.ID))
	metadata.Set("streamId", string(descriptor.StreamID))
	metadata.Set("sequenceNumber", int64(descriptor.SequenceNumber))
	metadata.Set("version", int64(descriptor.Version))
	metadata.Set("recordedAt", descriptor.RecordedAt)

	return event.NewWithMetadata(payload, metadata), nil
}

// StreamSliceToEventList converts a StreamSlice to an event.List.
func (c *EventConverter) StreamSliceToEventList(slice StreamSlice) (event.List, error) {
	var events event.List
	for _, d := range slice.Descriptors {
		e, err := c.FromRecordedEventDescriptor(d)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

// RegisterEvent registers an event and its type with this converter.
func (c *EventConverter) RegisterEvent(e event.Payload) *EventConverter {
	if _, found := c.events[e.TypeName()]; found {
		return c
	}

	typ := reflect.TypeOf(e)
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	c.events[e.TypeName()] = typ

	return c
}

// find a pointer to a struct of the event's type to be used for converting.
func (c *EventConverter) find(tn event.PayloadTypeName) (event.Payload, error) {
	evt, found := c.events[tn]
	if !found {
		return nil, errors.Errorf("no event registered for type name \"%s\"", tn)
	}

	// Create a new instance
	ret := reflect.ValueOf(reflect.New(evt).Interface()).Elem().Interface().(event.Payload)

	// Create a pointer since most decoders require pointer values.
	evtPtr := reflect.New(reflect.TypeOf(ret)).Interface().(event.Payload)
	return evtPtr, nil
}
