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
	"encoding/json"
	"github.com/jwillp/go-system/misas"
	"github.com/jwillp/go-system/misas/clock"
	"github.com/pkg/errors"
	"time"
)

// Descriptor Describes a Prediction allowing to save additional data alongside it.
type Descriptor struct {
	ID          ID
	TypeName    TypeName
	Payload     Payload
	Metadata    misas.Metadata
	PredictedAt time.Time
	WillOccurAt time.Time
}

// Payload simple data structure representing the data of a Prediction as read from some storage
type Payload map[string]any

// ValueAt returns a value of a field.
func (p Payload) ValueAt(fieldName string) any {
	if v, found := p[fieldName]; found {
		return v
	}

	return nil
}

// HasField indicates if the Payload contains a certain field or not.
func (p Payload) HasField(fieldName string) bool {
	_, found := p[fieldName]
	return found
}

// Store System responsible for durably storing predictions and allowing their retrieval.
type Store interface {
	// Add a prediction to this store.
	Add(p Prediction, m misas.Metadata) error

	// Remove a prediction from this store.
	Remove(id ID) error

	// FindOccurredBefore returns all the predictions in the store that have occurred before a certain date.
	FindOccurredBefore(dt time.Time) ([]Descriptor, error)
}

// InMemoryStore is an implementation of a prediction.Store that stores predictions in memory.
type InMemoryStore struct {
	Descriptors map[ID]Descriptor
	clock       clock.Clock
}

func NewInMemoryStore(c clock.Clock) *InMemoryStore {
	return &InMemoryStore{clock: c, Descriptors: map[ID]Descriptor{}}
}

func (i *InMemoryStore) Add(p Prediction, m misas.Metadata) error {
	marshal, err := json.Marshal(p)
	if err != nil {
		return errors.Wrapf(err, "could not add prediction %v", p)
	}

	var payload map[string]any
	if err := json.Unmarshal(marshal, &payload); err != nil {
		return errors.Wrapf(err, "could not add prediction %v", p)
	}

	d := Descriptor{
		ID:          p.ID(),
		TypeName:    p.TypeName(),
		Payload:     payload,
		Metadata:    m,
		PredictedAt: i.clock.Now(),
		WillOccurAt: p.WillOccurAt(),
	}
	i.Descriptors[d.ID] = d
	return nil
}

func (i *InMemoryStore) Remove(id ID) error {
	delete(i.Descriptors, id)
	return nil
}

func (i *InMemoryStore) FindOccurredBefore(dt time.Time) ([]Descriptor, error) {
	var result []Descriptor
	for _, descriptor := range i.Descriptors {
		willOccurAt := descriptor.WillOccurAt
		if willOccurAt.Equal(dt) || willOccurAt.Before(dt) {
			result = append(result, descriptor)
		}
	}

	return result, nil
}
