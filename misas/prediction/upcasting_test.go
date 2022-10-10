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
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

type upcasterA struct {
}

func (u upcasterA) supports(UpcastableDescriptor) bool { return true }

func (u upcasterA) upcast(descriptor UpcastableDescriptor) []UpcastableDescriptor {
	parts := strings.Split(descriptor.Payload.ValueAt("fullName").(string), " ")
	firstName := parts[0]
	lastName := parts[1]

	return []UpcastableDescriptor{
		descriptor.WithPayload(
			descriptor.Payload.
				withFieldAdded("firstName", firstName).
				withFieldAdded("lastName", lastName).
				WithFieldRemoved("fullName"),
		),
	}
}

type upcasterB struct{}

func (u upcasterB) supports(UpcastableDescriptor) bool { return true }

func (u upcasterB) upcast(descriptor UpcastableDescriptor) []UpcastableDescriptor {
	// Split into multiple descriptors
	return []UpcastableDescriptor{
		descriptor.
			withID(descriptor.ID + "fn").
			WithTypeName("user.first_name_changed").
			WithPayload(UpcastablePayload{"firstName": descriptor.Payload.ValueAt("firstName")}),

		descriptor.
			withID(descriptor.ID + "ln").
			WithTypeName("user.last_name_changed").
			WithPayload(UpcastablePayload{"lastName": descriptor.Payload.ValueAt("lastName")}),
	}
}

func TestEventUpcasterChain_doUpcast(t *testing.T) {
	chain := UpcasterChain{upcasters: []Upcaster{
		upcasterA{},
		upcasterB{},
	}}

	descriptor := UpcastableDescriptor{
		ID:          "event#1",
		TypeName:    "user.full_name_changed",
		Payload:     UpcastablePayload{"fullName": "Jane Doe"},
		PredictedAt: time.Time{},
		Metadata:    UpcastableMetadata{},
	}

	assert.True(t, chain.supports(descriptor))

	events := chain.upcast(descriptor)
	assert.Len(t, events, 2)
	assert.Equal(t, UpcastableDescriptor{
		ID:          "event#1fn",
		TypeName:    "user.first_name_changed",
		Payload:     UpcastablePayload{"firstName": "Jane"},
		PredictedAt: time.Time{},
		Metadata:    UpcastableMetadata{},
	}, events[0])

	assert.Equal(t, UpcastableDescriptor{
		ID:          "event#1ln",
		TypeName:    "user.last_name_changed",
		Payload:     UpcastablePayload{"lastName": "Doe"},
		PredictedAt: time.Time{},
		Metadata:    UpcastableMetadata{},
	}, events[1])
}
