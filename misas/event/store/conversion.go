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
	"fmt"
	"github.com/google/uuid"
	"github.com/morebec/go-errors/errors"
	"github.com/morebec/misas-go/misas/event"
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
	ec.RegisterEventPayload(StreamTruncatedEvent{})
	return ec
}

const EventConversionErrorCode = "event_conversion_failed"

// ConvertEventToDescriptor converts an event.Event to an DescriptorPayload to be used with an EventDescriptor.
func (c *EventConverter) ConvertEventToDescriptor(evt event.Event) (EventDescriptor, error) {
	payload, err := c.ConvertEventPayloadToDescriptorPayload(evt.Payload)
	if err != nil {
		return EventDescriptor{}, err
	}

	descriptor := EventDescriptor{
		ID:       EventID(uuid.NewString()),
		TypeName: evt.Payload.TypeName(),
		Payload:  payload,
		Metadata: nil,
	}

	return descriptor, nil
}

// ConvertEventPayloadToDescriptorPayload converts an event.Payload to a DescriptorPayload
func (c *EventConverter) ConvertEventPayloadToDescriptorPayload(p event.Payload) (DescriptorPayload, error) {

	// attempt registering if it was not already done before
	c.RegisterEventPayload(p)

	marshal, err := json.Marshal(p)
	if err != nil {
		return nil, errors.WrapWithMessage(err, EventConversionErrorCode, fmt.Sprintf(
			"failed converting event \"%s\" to DescriptorPayload",
			p.TypeName(),
		))
	}

	var payload DescriptorPayload
	if err := json.Unmarshal(marshal, &payload); err != nil {
		return nil, errors.WrapWithMessage(err, EventConversionErrorCode, fmt.Sprintf(
			"failed converting event \"%s\" to DescriptorPayload",
			p.TypeName(),
		))
	}

	return payload, nil
}

// ConvertEventListToDescriptorSlice converts a list of event.Event to a list of EventDescriptor.
func (c *EventConverter) ConvertEventListToDescriptorSlice(l event.List) []EventDescriptor {
	var descriptors []EventDescriptor
	for _, e := range l {
		d, err := c.ConvertEventToDescriptor(e)
		if err != nil {
			return nil
		}
		descriptors = append(descriptors, d)
	}

	return descriptors
}

// ConvertDescriptorToEvent loads an event.Event from a RecordedEventDescriptor
func (c *EventConverter) ConvertDescriptorToEvent(d RecordedEventDescriptor) (event.Event, error) {

	p, err := c.ConvertDescriptorPayloadToEventPayload(d.Payload, d.TypeName)
	if err != nil {
		return event.Event{}, errors.WrapWithMessage(
			err,
			EventConversionErrorCode,
			fmt.Sprintf("failed converting descriptor %s to %s", d.ID, d.TypeName),
		)
	}

	metadata := d.Metadata
	metadata.Set("id", string(d.ID))
	metadata.Set("streamId", string(d.StreamID))
	metadata.Set("sequenceNumber", int64(d.SequenceNumber))
	metadata.Set("version", int64(d.Version))
	metadata.Set("recordedAt", d.RecordedAt)

	return event.NewWithMetadata(p, metadata), nil
}

// ConvertDescriptorPayloadToEventPayload converts a DescriptorPayload to an event.Payload
func (c *EventConverter) ConvertDescriptorPayloadToEventPayload(dp DescriptorPayload, t event.PayloadTypeName) (event.Payload, error) {
	evt, err := c.findPayloadStruct(t)
	if err != nil {
		return nil, errors.WrapWithMessage(
			err,
			EventConversionErrorCode,
			fmt.Sprintf("failed converting descriptor to %s", t),
		)
	}

	marshal, err := json.Marshal(dp)
	if err != nil {
		return nil, errors.WrapWithMessage(
			err,
			EventConversionErrorCode,
			fmt.Sprintf("failed converting descriptor to %s", t),
		)
	}

	if err := json.Unmarshal(marshal, evt); err != nil {
		return nil, errors.WrapWithMessage(
			err,
			EventConversionErrorCode,
			fmt.Sprintf("failed converting descriptor to %s", t),
		)
	}

	// Convert *event.Event to event.Event
	payload := reflect.ValueOf(evt).Elem().Interface().(event.Payload)

	return payload, nil
}

// ConvertStreamSliceToEventList converts a StreamSlice to an event.List.
func (c *EventConverter) ConvertStreamSliceToEventList(slice StreamSlice) (event.List, error) {
	var events event.List
	for _, d := range slice.Descriptors {
		e, err := c.ConvertDescriptorToEvent(d)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

// RegisterEventPayload registers an event and its type with this converter.
func (c *EventConverter) RegisterEventPayload(e event.Payload) *EventConverter {
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

// findPayloadStruct a pointer to a struct of the event's type to be used for converting.
func (c *EventConverter) findPayloadStruct(tn event.PayloadTypeName) (event.Payload, error) {
	evt, found := c.events[tn]
	if !found {
		return nil, errors.NewWithMessage(
			EventConversionErrorCode,
			fmt.Sprintf("no event registered for type name \"%s\"", tn),
		)
	}

	// Create a new instance
	ret := reflect.ValueOf(reflect.New(evt).Interface()).Elem().Interface().(event.Payload)

	// Create a pointer since most decoders require pointer values.
	evtPtr := reflect.New(reflect.TypeOf(ret)).Interface().(event.Payload)
	return evtPtr, nil
}
