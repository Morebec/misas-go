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

package prediction

import (
	"github.com/morebec/go-system/misas"
	"time"
)

// UpcastableDescriptor Describes a Prediction which can be upcasted to a new form. This is intended to be use by the store to be able to load a Prediction
// into the system that might have undergone some schema changes between the time it was saved in the Store and the time it should occur.
type UpcastableDescriptor struct {
	ID          ID
	TypeName    TypeName
	Payload     UpcastablePayload
	PredictedAt time.Time
	WillOccurAt time.Time
	Metadata    UpcastableMetadata
}

func NewUpcastableDescriptorFromDescriptor(d Descriptor) UpcastableDescriptor {
	return UpcastableDescriptor{
		ID:          d.ID,
		TypeName:    d.TypeName,
		Payload:     UpcastablePayload(d.Payload),
		PredictedAt: d.PredictedAt,
		WillOccurAt: d.WillOccurAt,
		Metadata:    UpcastableMetadata(d.Metadata),
	}
}

// withMetadata Returns a copy of this UpcastableDescriptor with the provided metadata.
func (d UpcastableDescriptor) withMetadata(metadata UpcastableMetadata) UpcastableDescriptor {
	d.Metadata = metadata
	return d
}

// WithPayload Returns a copy of this UpcastableDescriptor with the provided Payload.
func (d UpcastableDescriptor) WithPayload(p UpcastablePayload) UpcastableDescriptor {
	d.Payload = UpcastablePayload(p)
	return d
}

// WithTypeName returns a copy of this UpcastableDescriptor with the type name renamed.
func (d UpcastableDescriptor) WithTypeName(n TypeName) UpcastableDescriptor {
	d.TypeName = n
	return d
}

func (d UpcastableDescriptor) withID(id ID) UpcastableDescriptor {
	d.ID = id
	return d
}

func (d UpcastableDescriptor) ToDescriptor() Descriptor {
	return Descriptor{
		ID:          d.ID,
		TypeName:    d.TypeName,
		Payload:     Payload(d.Payload),
		Metadata:    misas.Metadata(d.Metadata),
		PredictedAt: d.PredictedAt,
		WillOccurAt: d.WillOccurAt,
	}
}

// UpcastableMetadata simple data structure representing the data of an UpcastableMetadata as read from some storage, allowing simple operations on it.
type UpcastableMetadata misas.Metadata

// WithKey returns a copy of this UpcastableMetadata with a key renamed.
func (m UpcastableMetadata) WithKey(k string, defaultValue any) UpcastableMetadata {
	m[k] = defaultValue
	return m
}

// WithKeyRenamed returns a copy of this UpcastableMetadata with a key renamed.
func (m UpcastableMetadata) WithKeyRenamed(k string, newName string) UpcastableMetadata {
	return m.WithKey(newName, m.ValueAt(k)).WithKeyRemoved(k)
}

// WithKeyRemoved Returns a copy of this UpcastableMetadata with a key removed.
func (m UpcastableMetadata) WithKeyRemoved(k string) UpcastableMetadata {
	delete(m, k)
	return m
}

// WithValueUpdated Returns a copy of this UpcastableMetadata with the value of a certain key updated to a new value.
func (m UpcastableMetadata) WithValueUpdated(k string, value any) UpcastableMetadata {
	m[k] = value
	return m
}

// ValueAt returns a value of a key.
func (m UpcastableMetadata) ValueAt(k string) any {
	if v, found := m[k]; found {
		return v
	}

	return nil
}

func (m UpcastableMetadata) HasKey(k string) bool {
	_, found := m[k]
	return found
}

// UpcastablePayload is an implementation of Payload allowing simple operations on it for upcasting purposes.
type UpcastablePayload Payload

// ValueAt returns a value of a field.
func (u UpcastablePayload) ValueAt(fieldName string) any {
	if v, found := u[fieldName]; found {
		return v
	}

	return nil
}

// HasField indicates if the UpcastablePayload contains a certain field or not.
func (u UpcastablePayload) HasField(fieldName string) bool {
	_, found := u[fieldName]
	return found
}

// withFieldAdded returns a copy of this Payload with a field renamed.
func (u UpcastablePayload) withFieldAdded(fieldName string, defaultValue any) UpcastablePayload {
	u[fieldName] = defaultValue
	return u
}

// WithFieldRenamed returns a copy of this UpcastablePayload with a field renamed.
func (u UpcastablePayload) WithFieldRenamed(fieldName string, newName string) UpcastablePayload {
	return u.withFieldAdded(newName, u.ValueAt(fieldName)).WithFieldRemoved(fieldName)
}

// WithFieldRemoved Returns a copy of this UpcastablePayload with a field removed.
func (u UpcastablePayload) WithFieldRemoved(fieldName string) UpcastablePayload {
	delete(u, fieldName)
	return u
}

// WithFieldValueUpdated Returns a copy of this UpcastablePayload with the value of a certain field updated to a new value.
func (u UpcastablePayload) WithFieldValueUpdated(fieldName string, value any) UpcastablePayload {
	u[fieldName] = value
	return u
}

type Upcaster interface {
	// Indicates if this upcaster supports a given TypeName
	supports(descriptor UpcastableDescriptor) bool

	upcast(descriptor UpcastableDescriptor) []UpcastableDescriptor
}

// UpcasterChain Allows to pass an UpcastableDescriptor through a series of upcasters.
// It works like a pipeline passing the result of the previous upcaster to the next one.
type UpcasterChain struct {
	upcasters []Upcaster
}

func (c *UpcasterChain) supports(descriptor UpcastableDescriptor) bool {
	for _, u := range c.upcasters {
		if u.supports(descriptor) {
			return true
		}
	}

	return false
}

func (c *UpcasterChain) upcast(descriptor UpcastableDescriptor) []UpcastableDescriptor {
	return c.doUpcast(c.upcasters, descriptor)
}

func (c *UpcasterChain) doUpcast(upcasters []Upcaster, descriptor UpcastableDescriptor) []UpcastableDescriptor {
	if len(upcasters) == 0 {
		return []UpcastableDescriptor{descriptor}
	}

	head := upcasters[0:1]
	tail := upcasters[1:]

	upcaster := head[0]

	var descriptors []UpcastableDescriptor
	if upcaster.supports(descriptor) {
		descriptors = upcaster.upcast(descriptor)
	}

	var result []UpcastableDescriptor
	for _, d := range descriptors {
		if upcaster.supports(d) {
			result = append(result, c.doUpcast(tail, d)...)
		}
	}

	return result
}

func (c *UpcasterChain) AddUpcasters(u ...Upcaster) *UpcasterChain {
	c.upcasters = append(c.upcasters, u...)
	return c
}

type UpcastingPredictionStoreDecorator struct {
	chain *UpcasterChain
	inner Store
}

func NewUpcastingPredictionStoreDecorator(inner Store, chain *UpcasterChain) UpcastingPredictionStoreDecorator {
	return UpcastingPredictionStoreDecorator{inner: inner, chain: chain}
}

func (u UpcastingPredictionStoreDecorator) Add(p Prediction, m misas.Metadata) error {
	return u.inner.Add(p, m)
}

func (u UpcastingPredictionStoreDecorator) Remove(id ID) error {
	return u.inner.Remove(id)
}

func (u UpcastingPredictionStoreDecorator) FindOccurredBefore(dt time.Time) ([]Descriptor, error) {
	loaded, err := u.inner.FindOccurredBefore(dt)
	if err != nil {
		return nil, err
	}

	var out []Descriptor
	for _, d := range loaded {
		upcastable := NewUpcastableDescriptorFromDescriptor(d)
		if u.chain.supports(upcastable) {
			upcasted := u.chain.upcast(upcastable)
			for _, up := range upcasted {
				out = append(out, up.ToDescriptor())
			}
		}
	}

	return out, nil
}
