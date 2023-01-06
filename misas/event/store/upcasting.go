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
	"github.com/morebec/misas-go/misas"
	"github.com/morebec/misas-go/misas/event"
	"time"
)

// UpcastableEventDescriptor Describes an Event which can be upcasted to a new form. This is intended to be use by the store to be able to load an Event
// into the system that might have undergone some schema changes since the time it was saved in the store.
// This data structure is intended to be immutable, therefore its fields should never be changed directly, instead ,use the methods
// it provides to obtain a modified copy.
type UpcastableEventDescriptor struct {
	ID             EventID
	TypeName       event.PayloadTypeName
	Payload        UpcastableEventPayload
	StreamID       StreamID
	Version        StreamVersion
	SequenceNumber SequenceNumber
	RecordedAt     time.Time
	Metadata       UpcastableEventMetadata
}

func newUpcastableEventDescriptorFromRecordedEventDescriptor(descriptor RecordedEventDescriptor) UpcastableEventDescriptor {
	return UpcastableEventDescriptor{
		ID:             descriptor.ID,
		TypeName:       descriptor.TypeName,
		Payload:        UpcastableEventPayload(descriptor.Payload),
		StreamID:       descriptor.StreamID,
		Version:        descriptor.Version,
		SequenceNumber: descriptor.SequenceNumber,
		RecordedAt:     descriptor.RecordedAt,
		Metadata:       UpcastableEventMetadata(descriptor.Metadata),
	}
}

// withMetadata Returns a copy of this UpcastableEventDescriptor with the provided metadata.
func (d UpcastableEventDescriptor) withMetadata(metadata UpcastableEventMetadata) UpcastableEventDescriptor {
	d.Metadata = metadata
	return d
}

// WithPayload Returns a copy of this UpcastableEventDescriptor with the provided event data.
func (d UpcastableEventDescriptor) WithPayload(payload UpcastableEventPayload) UpcastableEventDescriptor {
	d.Payload = payload
	return d
}

// WithTypeName returns a copy of this event with the type name renamed.
func (d UpcastableEventDescriptor) WithTypeName(n event.PayloadTypeName) UpcastableEventDescriptor {
	d.TypeName = n
	return d
}

func (d UpcastableEventDescriptor) WithID(id EventID) UpcastableEventDescriptor {
	d.ID = id
	return d
}

// ToRecordedEventDescriptor Converts this UpcastableEventDescriptor to a RecordedEventDescriptor
func (d UpcastableEventDescriptor) ToRecordedEventDescriptor() RecordedEventDescriptor {
	return RecordedEventDescriptor{
		ID:             d.ID,
		TypeName:       d.TypeName,
		Payload:        DescriptorPayload(d.Payload),
		Metadata:       misas.Metadata(d.Metadata),
		StreamID:       d.StreamID,
		Version:        d.Version,
		SequenceNumber: d.SequenceNumber,
		RecordedAt:     d.RecordedAt,
	}
}

// UpcastableEventMetadata simple data structure representing the data of an event's Metadata as read from some storage, allowing simple operations on it.
type UpcastableEventMetadata misas.Metadata

// WithValue returns a copy of this UpcastableEventMetadata with a value for a given key.
// This method can be useful to provide a default value on events.
func (m UpcastableEventMetadata) WithValue(k string, value any) UpcastableEventMetadata {
	nm := m
	if nm == nil {
		nm = UpcastableEventMetadata{}
	}
	nm[k] = value
	return nm
}

// WithKeyRenamed returns a copy of this UpcastableEventMetadata with a key renamed.
func (m UpcastableEventMetadata) WithKeyRenamed(k string, newName string) UpcastableEventMetadata {
	if m == nil {
		return m
	}

	return m.WithValue(newName, m.ValueAt(k, nil)).WithKeyRemoved(k)
}

// WithKeyRemoved Returns a copy of this UpcastableEventMetadata with a key removed.
func (m UpcastableEventMetadata) WithKeyRemoved(k string) UpcastableEventMetadata {
	delete(m, k)
	return m
}

// WithValueUpdated Returns a copy of this UpcastableEventMetadata with the value of a certain key updated to a new value.
// If the key does not exist, no change will be returned.
func (m UpcastableEventMetadata) WithValueUpdated(k string, value any) UpcastableEventMetadata {
	if m.HasKey(k) == false {
		return m
	}

	return m.WithValue(k, value)
}

// ValueAt returns a value of a key or a default value if it was not found.
func (m UpcastableEventMetadata) ValueAt(k string, defaultValue any) any {
	if v, found := m[k]; found {
		return v
	}

	return defaultValue
}

// HasKey Indicates if a value exists at a given key.
func (m UpcastableEventMetadata) HasKey(k string) bool {
	_, found := m[k]
	return found
}

// UpcastableEventPayload simple data structure representing the data of an event as read from some event storage, allowing simple operations on it.
type UpcastableEventPayload map[string]any

// withFieldAdded returns a copy of this UpcastableEventPayload with a field renamed. If the field is already defined, its value will not be overwritten.
// instead be explicit and use methods such as HasField and withFieldUpdated.
func (p UpcastableEventPayload) withFieldAdded(fieldName string, defaultValue any) UpcastableEventPayload {
	// Do not overwrite if already exists.
	if p.HasField(fieldName) {
		return p
	}

	p[fieldName] = defaultValue
	return p
}

// WithFieldRenamed returns a copy of this UpcastableEventPayload with a field renamed.
func (p UpcastableEventPayload) WithFieldRenamed(fieldName string, newName string) UpcastableEventPayload {
	if p == nil {
		return p
	}

	if !p.HasField(fieldName) {
		return p
	}

	return p.withFieldAdded(newName, p.ValueAt(fieldName, nil)).WithFieldRemoved(fieldName)
}

// WithFieldRemoved Returns a copy of this UpcastableEventPayload with a field removed.
func (p UpcastableEventPayload) WithFieldRemoved(fieldName string) UpcastableEventPayload {
	delete(p, fieldName)
	return p
}

// WithFieldValueUpdated Returns a copy of this UpcastableEventPayload with the value of a certain field updated to a new value.
func (p UpcastableEventPayload) WithFieldValueUpdated(fieldName string, value any) UpcastableEventPayload {
	if p == nil {
		return p
	}

	if !p.HasField(fieldName) {
		return p
	}

	p[fieldName] = value
	return p
}

// ValueAt returns a value of a field.
func (p UpcastableEventPayload) ValueAt(fieldName string, defaultValue any) any {
	if v, found := p[fieldName]; found {
		return v
	}

	return defaultValue
}

// HasField indicates if the UpcastableEventPayload contains a certain field or not.
func (p UpcastableEventPayload) HasField(fieldName string) bool {
	_, found := p[fieldName]
	return found
}

// Upcaster An upcaster is a service responsible for updating the schema of an event to better represent what is implemented in code.
type Upcaster interface {
	// Supports indicates if this upcaster supports a given UpcastableEventDescriptor
	Supports(descriptor UpcastableEventDescriptor) bool

	Upcast(descriptor UpcastableEventDescriptor) []UpcastableEventDescriptor
}

// UpcasterChain Allows to pass an event descriptor through a series of upcasters.
// It works like a pipeline passing the result of the previous upcaster to the next one.
type UpcasterChain struct {
	upcasters []Upcaster
}

func NewUpcasterChain(upcasters ...Upcaster) *UpcasterChain {
	return &UpcasterChain{upcasters: upcasters}
}

func (c *UpcasterChain) Supports(descriptor UpcastableEventDescriptor) bool {
	for _, u := range c.upcasters {
		if u.Supports(descriptor) {
			return true
		}
	}

	return false
}

func (c *UpcasterChain) Upcast(descriptor UpcastableEventDescriptor) []UpcastableEventDescriptor {
	return c.doUpcast(c.upcasters, descriptor)
}

func (c *UpcasterChain) doUpcast(upcasters []Upcaster, descriptor UpcastableEventDescriptor) []UpcastableEventDescriptor {
	if len(upcasters) == 0 {
		return []UpcastableEventDescriptor{descriptor}
	}

	head := upcasters[0:1]
	tail := upcasters[1:]

	upcaster := head[0]
	var descriptors []UpcastableEventDescriptor
	if upcaster.Supports(descriptor) {
		descriptors = upcaster.Upcast(descriptor)
	}

	var result []UpcastableEventDescriptor
	for _, d := range descriptors {
		result = append(result, c.doUpcast(tail, d)...)
	}
	return result
}

func (c *UpcasterChain) AddUpcasters(u ...Upcaster) *UpcasterChain {
	c.upcasters = append(c.upcasters, u...)
	return c
}

// UpcasterFunc allows defining upcasters using functions instead of a concrete type.
type UpcasterFunc func() (
	func(descriptor UpcastableEventDescriptor) bool,
	func(descriptor UpcastableEventDescriptor) []UpcastableEventDescriptor,
)

func (u UpcasterFunc) Supports(descriptor UpcastableEventDescriptor) bool {
	supports, _ := u()
	return supports(descriptor)
}

func (u UpcasterFunc) Upcast(descriptor UpcastableEventDescriptor) []UpcastableEventDescriptor {
	_, upcast := u()
	return upcast(descriptor)
}

// UpcastingEventStoreDecorator decorator around an event store that allows upcasting events using an upcaster chain.
type UpcastingEventStoreDecorator struct {
	chain *UpcasterChain
	inner EventStore
}

// NewUpcastingEventStoreDecorator returns a new upcasting event store decorator.
func NewUpcastingEventStoreDecorator(inner EventStore, chain *UpcasterChain) *UpcastingEventStoreDecorator {
	return &UpcastingEventStoreDecorator{inner: inner, chain: chain}
}

func (u UpcastingEventStoreDecorator) GlobalStreamID() StreamID {
	return u.inner.GlobalStreamID()
}

func (u UpcastingEventStoreDecorator) AppendToStream(ctx context.Context, streamID StreamID, events []EventDescriptor, opts ...AppendToStreamOption) error {
	return u.inner.AppendToStream(ctx, streamID, events, opts...)
}

func (u UpcastingEventStoreDecorator) ReadFromStream(ctx context.Context, streamID StreamID, opts ...ReadFromStreamOption) (StreamSlice, error) {
	stream, err := u.inner.ReadFromStream(ctx, streamID, opts...)
	if err != nil {
		return StreamSlice{}, err
	}

	upcastedSlice := StreamSlice{
		StreamID:    streamID,
		Descriptors: []RecordedEventDescriptor{},
	}

	for _, d := range stream.Descriptors {
		upcastable := newUpcastableEventDescriptorFromRecordedEventDescriptor(d)
		if !u.chain.Supports(upcastable) {
			upcastedSlice.Descriptors = append(upcastedSlice.Descriptors, d)
			continue
		}
		upcastedEvents := u.chain.Upcast(upcastable)
		for _, up := range upcastedEvents {
			upcastedSlice.Descriptors = append(upcastedSlice.Descriptors, up.ToRecordedEventDescriptor())
		}
	}

	return upcastedSlice, nil
}

func (u UpcastingEventStoreDecorator) TruncateStream(ctx context.Context, streamID StreamID, opts ...TruncateStreamOption) error {
	return u.inner.TruncateStream(ctx, streamID, opts...)
}

func (u UpcastingEventStoreDecorator) DeleteStream(ctx context.Context, id StreamID) error {
	return u.inner.DeleteStream(ctx, id)
}

func (u UpcastingEventStoreDecorator) SubscribeToStream(ctx context.Context, streamID StreamID, opts ...SubscribeToStreamOption) (Subscription, error) {
	return u.inner.SubscribeToStream(ctx, streamID, opts...)
}

func (u UpcastingEventStoreDecorator) StreamExists(ctx context.Context, id StreamID) (bool, error) {
	return u.inner.StreamExists(ctx, id)
}

func (u UpcastingEventStoreDecorator) GetStream(ctx context.Context, id StreamID) (Stream, error) {
	return u.inner.GetStream(ctx, id)
}

func (u UpcastingEventStoreDecorator) Clear(ctx context.Context) error {
	return u.inner.Clear(ctx)
}
