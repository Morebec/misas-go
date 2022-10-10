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
	"github.com/google/uuid"
	"github.com/jwillp/go-system/misas"
	"github.com/jwillp/go-system/misas/clock"
	"github.com/jwillp/go-system/misas/event"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func TestUpcastableEventDescriptor_ToRecordedEventDescriptor(t *testing.T) {
	type fields struct {
		ID             EventID
		TypeName       event.TypeName
		Payload        UpcastableEventPayload
		StreamID       StreamID
		Version        StreamVersion
		SequenceNumber SequenceNumber
		RecordedAt     time.Time
		Metadata       UpcastableEventMetadata
	}
	tests := []struct {
		name   string
		fields fields
		want   RecordedEventDescriptor
	}{
		{
			name: "",
			fields: fields{
				ID:       "#000",
				TypeName: "unit.test",
				Payload: UpcastableEventPayload{
					"hello": "world",
				},
				StreamID:       "unit.test",
				Version:        10,
				SequenceNumber: 15,
				RecordedAt:     time.Date(2020, time.January, 01, 00, 00, 00, 00, time.UTC),
				Metadata: UpcastableEventMetadata{
					"meta": "data",
				},
			},
			want: RecordedEventDescriptor{
				ID:       "#000",
				TypeName: "unit.test",
				Payload: EventPayload{
					"hello": "world",
				},
				Metadata: misas.Metadata{
					"meta": "data",
				},
				StreamID:       "unit.test",
				Version:        10,
				SequenceNumber: 15,
				RecordedAt:     time.Date(2020, time.January, 01, 00, 00, 00, 00, time.UTC),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := UpcastableEventDescriptor{
				ID:             tt.fields.ID,
				TypeName:       tt.fields.TypeName,
				Payload:        tt.fields.Payload,
				StreamID:       tt.fields.StreamID,
				Version:        tt.fields.Version,
				SequenceNumber: tt.fields.SequenceNumber,
				RecordedAt:     tt.fields.RecordedAt,
				Metadata:       tt.fields.Metadata,
			}
			assert.Equalf(t, tt.want, d.ToRecordedEventDescriptor(), "ToRecordedEventDescriptor()")
		})
	}
}

func TestUpcastableEventDescriptor_WithID(t *testing.T) {
	type fields struct {
		ID             EventID
		TypeName       event.TypeName
		Payload        UpcastableEventPayload
		StreamID       StreamID
		Version        StreamVersion
		SequenceNumber SequenceNumber
		RecordedAt     time.Time
		Metadata       UpcastableEventMetadata
	}
	type args struct {
		id EventID
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   UpcastableEventDescriptor
	}{
		{
			name: "",
			fields: fields{
				ID:       "#000",
				TypeName: "unit.test",
				Payload: UpcastableEventPayload{
					"hello": "world",
				},
				StreamID:       "unit.test",
				Version:        10,
				SequenceNumber: 15,
				RecordedAt:     time.Date(2020, time.January, 01, 00, 00, 00, 00, time.UTC),
				Metadata: UpcastableEventMetadata{
					"meta": "data",
				},
			},
			args: args{
				id: "#150",
			},
			want: UpcastableEventDescriptor{
				ID:       "#150",
				TypeName: "unit.test",
				Payload: UpcastableEventPayload{
					"hello": "world",
				},
				StreamID:       "unit.test",
				Version:        10,
				SequenceNumber: 15,
				RecordedAt:     time.Date(2020, time.January, 01, 00, 00, 00, 00, time.UTC),
				Metadata: UpcastableEventMetadata{
					"meta": "data",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := UpcastableEventDescriptor{
				ID:             tt.fields.ID,
				TypeName:       tt.fields.TypeName,
				Payload:        tt.fields.Payload,
				StreamID:       tt.fields.StreamID,
				Version:        tt.fields.Version,
				SequenceNumber: tt.fields.SequenceNumber,
				RecordedAt:     tt.fields.RecordedAt,
				Metadata:       tt.fields.Metadata,
			}
			assert.Equalf(t, tt.want, d.WithID(tt.args.id), "WithID(%v)", tt.args.id)
		})
	}
}

func TestUpcastableEventDescriptor_WithPayload(t *testing.T) {
	type fields struct {
		ID             EventID
		TypeName       event.TypeName
		Payload        UpcastableEventPayload
		StreamID       StreamID
		Version        StreamVersion
		SequenceNumber SequenceNumber
		RecordedAt     time.Time
		Metadata       UpcastableEventMetadata
	}
	type args struct {
		payload UpcastableEventPayload
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   UpcastableEventDescriptor
	}{
		{
			name: "",
			fields: fields{
				ID:       "#000",
				TypeName: "unit.test",
				Payload: UpcastableEventPayload{
					"hello": "world",
				},
				StreamID:       "unit.test",
				Version:        10,
				SequenceNumber: 15,
				RecordedAt:     time.Date(2020, time.January, 01, 00, 00, 00, 00, time.UTC),
				Metadata: UpcastableEventMetadata{
					"meta": "data",
				},
			},
			args: args{
				payload: UpcastableEventPayload{
					"hello": "unit test",
				},
			},
			want: UpcastableEventDescriptor{
				ID:       "#000",
				TypeName: "unit.test",
				Payload: UpcastableEventPayload{
					"hello": "unit test",
				},
				StreamID:       "unit.test",
				Version:        10,
				SequenceNumber: 15,
				RecordedAt:     time.Date(2020, time.January, 01, 00, 00, 00, 00, time.UTC),
				Metadata: UpcastableEventMetadata{
					"meta": "data",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := UpcastableEventDescriptor{
				ID:             tt.fields.ID,
				TypeName:       tt.fields.TypeName,
				Payload:        tt.fields.Payload,
				StreamID:       tt.fields.StreamID,
				Version:        tt.fields.Version,
				SequenceNumber: tt.fields.SequenceNumber,
				RecordedAt:     tt.fields.RecordedAt,
				Metadata:       tt.fields.Metadata,
			}
			assert.Equalf(t, tt.want, d.WithPayload(tt.args.payload), "WithPayload(%v)", tt.args.payload)
		})
	}
}

func TestUpcastableEventDescriptor_WithTypeName(t *testing.T) {
	type fields struct {
		ID             EventID
		TypeName       event.TypeName
		Payload        UpcastableEventPayload
		StreamID       StreamID
		Version        StreamVersion
		SequenceNumber SequenceNumber
		RecordedAt     time.Time
		Metadata       UpcastableEventMetadata
	}
	type args struct {
		n event.TypeName
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   UpcastableEventDescriptor
	}{
		{
			name: "",
			fields: fields{
				ID:       "#000",
				TypeName: "unit.test",
				Payload: UpcastableEventPayload{
					"hello": "world",
				},
				StreamID:       "unit.test",
				Version:        10,
				SequenceNumber: 15,
				RecordedAt:     time.Date(2020, time.January, 01, 00, 00, 00, 00, time.UTC),
				Metadata: UpcastableEventMetadata{
					"meta": "data",
				},
			},
			args: args{
				n: "type_name.changed",
			},
			want: UpcastableEventDescriptor{
				ID:       "#000",
				TypeName: "type_name.changed",
				Payload: UpcastableEventPayload{
					"hello": "world",
				},
				StreamID:       "unit.test",
				Version:        10,
				SequenceNumber: 15,
				RecordedAt:     time.Date(2020, time.January, 01, 00, 00, 00, 00, time.UTC),
				Metadata: UpcastableEventMetadata{
					"meta": "data",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := UpcastableEventDescriptor{
				ID:             tt.fields.ID,
				TypeName:       tt.fields.TypeName,
				Payload:        tt.fields.Payload,
				StreamID:       tt.fields.StreamID,
				Version:        tt.fields.Version,
				SequenceNumber: tt.fields.SequenceNumber,
				RecordedAt:     tt.fields.RecordedAt,
				Metadata:       tt.fields.Metadata,
			}
			assert.Equalf(t, tt.want, d.WithTypeName(tt.args.n), "WithTypeName(%v)", tt.args.n)
		})
	}
}

func TestUpcastableEventDescriptor_withMetadata(t *testing.T) {
	type fields struct {
		ID             EventID
		TypeName       event.TypeName
		Payload        UpcastableEventPayload
		StreamID       StreamID
		Version        StreamVersion
		SequenceNumber SequenceNumber
		RecordedAt     time.Time
		Metadata       UpcastableEventMetadata
	}
	type args struct {
		metadata UpcastableEventMetadata
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   UpcastableEventDescriptor
	}{
		{
			name: "",
			fields: fields{
				TypeName: "unit.test",
				Payload: UpcastableEventPayload{
					"hello": "world",
				},
				StreamID:       "unit.test",
				Version:        10,
				SequenceNumber: 15,
				RecordedAt:     time.Date(2020, time.January, 01, 00, 00, 00, 00, time.UTC),
				Metadata: UpcastableEventMetadata{
					"meta": "data",
				},
			},
			args: args{
				metadata: UpcastableEventMetadata{
					"updated": "metadata",
				},
			},
			want: UpcastableEventDescriptor{
				TypeName: "unit.test",
				Payload: UpcastableEventPayload{
					"hello": "world",
				},
				StreamID:       "unit.test",
				Version:        10,
				SequenceNumber: 15,
				RecordedAt:     time.Date(2020, time.January, 01, 00, 00, 00, 00, time.UTC),
				Metadata: UpcastableEventMetadata{
					"updated": "metadata",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := UpcastableEventDescriptor{
				ID:             tt.fields.ID,
				TypeName:       tt.fields.TypeName,
				Payload:        tt.fields.Payload,
				StreamID:       tt.fields.StreamID,
				Version:        tt.fields.Version,
				SequenceNumber: tt.fields.SequenceNumber,
				RecordedAt:     tt.fields.RecordedAt,
				Metadata:       tt.fields.Metadata,
			}
			assert.Equalf(t, tt.want, d.withMetadata(tt.args.metadata), "withMetadata(%v)", tt.args.metadata)
		})
	}
}

func TestUpcastableEventMetadata_ValueAt(t *testing.T) {
	type args struct {
		k            string
		defaultValue any
	}
	tests := []struct {
		name string
		m    UpcastableEventMetadata
		args args
		want any
	}{
		{
			name: "find actual value",
			m: UpcastableEventMetadata{
				"aKey": "aValue",
			},
			args: args{
				k:            "aKey",
				defaultValue: "defaultValue",
			},
			want: "aValue",
		},
		{
			name: "find default value",
			m:    UpcastableEventMetadata{},
			args: args{
				k:            "aKey",
				defaultValue: "defaultValue",
			},
			want: "defaultValue",
		},
		{
			name: "on nil",
			m:    nil,
			args: args{
				k:            "aKey",
				defaultValue: "defaultValue",
			},
			want: "defaultValue",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.m.ValueAt(tt.args.k, tt.args.defaultValue), "ValueAt(%v, %v)", tt.args.k, tt.args.defaultValue)
		})
	}
}

func TestUpcastableEventMetadata_HasKey(t *testing.T) {
	type args struct {
		k string
	}
	tests := []struct {
		name string
		m    UpcastableEventMetadata
		args args
		want bool
	}{
		{
			name: "key exists",
			m: UpcastableEventMetadata{
				"key": "value",
			},
			args: args{
				k: "key",
			},
			want: true,
		},
		{
			name: "key does not exist",
			m:    UpcastableEventMetadata{},
			args: args{
				k: "does not exist",
			},
			want: false,
		},
		{
			name: "on nil",
			m:    nil,
			args: args{
				k: "aKey",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.m.HasKey(tt.args.k), "HasKey(%v)", tt.args.k)
		})
	}
}

func TestUpcastableEventMetadata_WithKeyRemoved(t *testing.T) {
	type args struct {
		k string
	}
	tests := []struct {
		name string
		m    UpcastableEventMetadata
		args args
		want UpcastableEventMetadata
	}{
		{
			name: "",
			m: UpcastableEventMetadata{
				"keyToRemove": "value",
				"keep":        "value",
			},
			args: args{
				k: "keyToRemove",
			},
			want: UpcastableEventMetadata{
				"keep": "value",
			},
		},
		{
			name: "on nil",
			m:    nil,
			args: args{
				k: "key",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.m.WithKeyRemoved(tt.args.k), "WithKeyRemoved(%v)", tt.args.k)
		})
	}
}

func TestUpcastableEventMetadata_WithKeyRenamed(t *testing.T) {
	type args struct {
		k       string
		newName string
	}
	tests := []struct {
		name string
		m    UpcastableEventMetadata
		args args
		want UpcastableEventMetadata
	}{
		{
			name: "",
			m: UpcastableEventMetadata{
				"keyToRename": "value",
				"keep":        "value",
			},
			args: args{
				k:       "keyToRename",
				newName: "newName",
			},
			want: UpcastableEventMetadata{
				"newName": "value",
				"keep":    "value",
			},
		},
		{
			name: "on nil",
			m:    nil,
			args: args{
				k: "key",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.m.WithKeyRenamed(tt.args.k, tt.args.newName), "WithKeyRenamed(%v, %v)", tt.args.k, tt.args.newName)
		})
	}
}

func TestUpcastableEventMetadata_WithValue(t *testing.T) {
	type args struct {
		k     string
		value any
	}
	tests := []struct {
		name string
		m    UpcastableEventMetadata
		args args
		want UpcastableEventMetadata
	}{
		{
			name: "",
			m: UpcastableEventMetadata{
				"key":  "value",
				"keep": "value",
			},
			args: args{
				k:     "key",
				value: "new_value",
			},
			want: UpcastableEventMetadata{
				"key":  "new_value",
				"keep": "value",
			},
		},
		{
			name: "on nil",
			m:    nil,
			args: args{
				k:     "key",
				value: "value",
			},
			want: UpcastableEventMetadata{
				"key": "value",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.m.WithValue(tt.args.k, tt.args.value), "WithValue(%v, %v)", tt.args.k, tt.args.value)
		})
	}
}

func TestUpcastableEventMetadata_WithValueUpdated(t *testing.T) {
	type args struct {
		k     string
		value any
	}
	tests := []struct {
		name string
		m    UpcastableEventMetadata
		args args
		want UpcastableEventMetadata
	}{
		{
			name: "",
			m: UpcastableEventMetadata{
				"key":  "value",
				"keep": "value",
			},
			args: args{
				k:     "key",
				value: "new_value",
			},
			want: UpcastableEventMetadata{
				"key":  "new_value",
				"keep": "value",
			},
		},
		{
			name: "non existing value should not cause any change",
			m: UpcastableEventMetadata{
				"key": "value",
			},
			args: args{
				k:     "404",
				value: "new_value",
			},
			want: UpcastableEventMetadata{
				"key": "value",
			},
		},
		{
			name: "on nil, should return nil",
			m:    nil,
			args: args{
				k:     "key",
				value: "value",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.m.WithValueUpdated(tt.args.k, tt.args.value), "WithValueUpdated(%v, %v)", tt.args.k, tt.args.value)
		})
	}
}

func TestUpcastableEventPayload_ValueAt(t *testing.T) {
	type args struct {
		fieldName    string
		defaultValue any
	}
	tests := []struct {
		name string
		d    UpcastableEventPayload
		args args
		want any
	}{
		{
			name: "existing field",
			d: UpcastableEventPayload{
				"field": "value",
			},
			args: args{
				fieldName:    "field",
				defaultValue: nil,
			},
			want: "value",
		},
		{
			name: "non existing field",
			d:    UpcastableEventPayload{},
			args: args{
				fieldName:    "non existing field",
				defaultValue: "defaultValue",
			},
			want: "defaultValue",
		},
		{
			name: "on nil",
			d:    nil,
			args: args{
				fieldName:    "field",
				defaultValue: "defaultValue",
			},
			want: "defaultValue",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.ValueAt(tt.args.fieldName, tt.args.defaultValue), "ValueAt(%v)", tt.args.fieldName)
		})
	}
}

func TestUpcastableEventPayload_HasField(t *testing.T) {
	type args struct {
		fieldName string
	}
	tests := []struct {
		name string
		d    UpcastableEventPayload
		args args
		want bool
	}{
		{
			name: "existing field",
			d: UpcastableEventPayload{
				"field": "value",
			},
			args: args{
				fieldName: "field",
			},
			want: true,
		},
		{
			name: "non existing field",
			d: UpcastableEventPayload{
				"field": "value",
			},
			args: args{
				fieldName: "404",
			},
			want: false,
		},
		{
			name: "on nil",
			d:    nil,
			args: args{
				fieldName: "field",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.HasField(tt.args.fieldName), "HasField(%v)", tt.args.fieldName)
		})
	}
}

func TestUpcastableEventPayload_WithFieldRemoved(t *testing.T) {
	type args struct {
		fieldName string
	}
	tests := []struct {
		name string
		d    UpcastableEventPayload
		args args
		want UpcastableEventPayload
	}{
		{
			name: "existing field",
			d: UpcastableEventPayload{
				"field":         "value",
				"fieldToRemove": "value",
			},
			args: args{
				fieldName: "fieldToRemove",
			},
			want: UpcastableEventPayload{
				"field": "value",
			},
		},
		{
			name: "non existing field should not cause any change",
			d: UpcastableEventPayload{
				"field": "value",
			},
			args: args{
				fieldName: "404",
			},
			want: UpcastableEventPayload{
				"field": "value",
			},
		},
		{
			name: "on nil",
			d:    nil,
			args: args{
				fieldName: "field",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.WithFieldRemoved(tt.args.fieldName), "WithFieldRemoved(%v)", tt.args.fieldName)
		})
	}
}

func TestUpcastableEventPayload_WithFieldRenamed(t *testing.T) {
	type args struct {
		fieldName string
		newName   string
	}
	tests := []struct {
		name string
		d    UpcastableEventPayload
		args args
		want UpcastableEventPayload
	}{
		{
			name: "existing field",
			d: UpcastableEventPayload{
				"field":        "value",
				"anotherField": "anotherValue",
			},
			args: args{
				fieldName: "field",
				newName:   "someField",
			},
			want: UpcastableEventPayload{
				"someField":    "value",
				"anotherField": "anotherValue",
			},
		},
		{
			name: "non existing field should not cause any change",
			d: UpcastableEventPayload{
				"field":        "value",
				"anotherField": "anotherValue",
			},
			args: args{
				fieldName: "404",
				newName:   "NotFound",
			},
			want: UpcastableEventPayload{
				"field":        "value",
				"anotherField": "anotherValue",
			},
		},
		{
			name: "on nil",
			d:    nil,
			args: args{
				fieldName: "field",
				newName:   "someField",
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.WithFieldRenamed(tt.args.fieldName, tt.args.newName), "WithFieldRenamed(%v, %v)", tt.args.fieldName, tt.args.newName)
		})
	}
}

func TestUpcastableEventPayload_WithFieldValueUpdated(t *testing.T) {
	type args struct {
		fieldName string
		value     any
	}
	tests := []struct {
		name string
		d    UpcastableEventPayload
		args args
		want UpcastableEventPayload
	}{
		{
			name: "existing field",
			d: UpcastableEventPayload{
				"field": "value",
			},
			args: args{
				fieldName: "field",
				value:     "new_value",
			},
			want: UpcastableEventPayload{
				"field": "new_value",
			},
		},
		{
			name: "non existing field should not cause any change",
			d: UpcastableEventPayload{
				"field": "value",
			},
			args: args{
				fieldName: "non existing field",
				value:     "new_value",
			},
			want: UpcastableEventPayload{
				"field": "value",
			},
		},
		{
			name: "on nil should return nil",
			d:    nil,
			args: args{
				fieldName: "field",
				value:     "value",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.WithFieldValueUpdated(tt.args.fieldName, tt.args.value), "WithFieldValueUpdated(%v, %v)", tt.args.fieldName, tt.args.value)
		})
	}
}

func TestUpcastableEventPayload_withFieldAdded(t *testing.T) {
	type args struct {
		fieldName    string
		defaultValue any
	}
	tests := []struct {
		name string
		d    UpcastableEventPayload
		args args
		want UpcastableEventPayload
	}{
		{
			name: "non existing field",
			d: UpcastableEventPayload{
				"field": "value",
			},
			args: args{
				fieldName:    "new_field",
				defaultValue: "new_value",
			},
			want: UpcastableEventPayload{
				"field":     "value",
				"new_field": "new_value",
			},
		},
		{
			name: "existing field should not cause any change",
			d: UpcastableEventPayload{
				"field": "value",
			},
			args: args{
				fieldName:    "field",
				defaultValue: "new_value",
			},
			want: UpcastableEventPayload{
				"field": "value",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.withFieldAdded(tt.args.fieldName, tt.args.defaultValue), "withFieldAdded(%v, %v)", tt.args.fieldName, tt.args.defaultValue)
		})
	}
}

type upcasterA struct {
}

func (u upcasterA) Supports(descriptor UpcastableEventDescriptor) bool { return true }

func (u upcasterA) Upcast(descriptor UpcastableEventDescriptor) []UpcastableEventDescriptor {
	parts := strings.Split(descriptor.Payload.ValueAt("fullName", nil).(string), " ")
	firstName := parts[0]
	lastName := parts[1]

	return []UpcastableEventDescriptor{
		descriptor.WithPayload(
			descriptor.Payload.
				withFieldAdded("firstName", firstName).
				withFieldAdded("lastName", lastName).
				WithFieldRemoved("fullName"),
		),
	}
}

type upcasterB struct{}

func (u upcasterB) Supports(descriptor UpcastableEventDescriptor) bool { return true }

func (u upcasterB) Upcast(descriptor UpcastableEventDescriptor) []UpcastableEventDescriptor {
	// Split into multiple descriptors
	return []UpcastableEventDescriptor{
		descriptor.
			WithID(descriptor.ID + "fn").
			WithTypeName("user.first_name_changed").
			WithPayload(UpcastableEventPayload{"firstName": descriptor.Payload.ValueAt("firstName", nil)}),

		descriptor.
			WithID(descriptor.ID + "ln").
			WithTypeName("user.last_name_changed").
			WithPayload(UpcastableEventPayload{"lastName": descriptor.Payload.ValueAt("lastName", nil)}),
	}
}

func TestUpcasterChain_doUpcast(t *testing.T) {
	chain := UpcasterChain{upcasters: []Upcaster{
		upcasterA{},
		upcasterB{},
	}}

	descriptor := UpcastableEventDescriptor{
		ID:             "event#1",
		TypeName:       "user.full_name_changed",
		Payload:        UpcastableEventPayload{"fullName": "Jane Doe"},
		StreamID:       "user/jane-doe",
		Version:        0,
		SequenceNumber: 0,
		RecordedAt:     time.Time{},
		Metadata:       UpcastableEventMetadata{},
	}

	assert.True(t, chain.Supports(descriptor))

	events := chain.Upcast(descriptor)
	assert.Len(t, events, 2)
	assert.Equal(t, UpcastableEventDescriptor{
		ID:             "event#1fn",
		TypeName:       "user.first_name_changed",
		Payload:        UpcastableEventPayload{"firstName": "Jane"},
		StreamID:       "user/jane-doe",
		Version:        0,
		SequenceNumber: 0,
		RecordedAt:     time.Time{},
		Metadata:       UpcastableEventMetadata{},
	}, events[0])

	assert.Equal(t, UpcastableEventDescriptor{
		ID:             "event#1ln",
		TypeName:       "user.last_name_changed",
		Payload:        UpcastableEventPayload{"lastName": "Doe"},
		StreamID:       "user/jane-doe",
		Version:        0,
		SequenceNumber: 0,
		RecordedAt:     time.Time{},
		Metadata:       UpcastableEventMetadata{},
	}, events[1])
}

func TestUpcasterChain_Supports(t *testing.T) {
	type fields struct {
		upcasters []Upcaster
	}
	type args struct {
		descriptor UpcastableEventDescriptor
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := UpcasterChain{
				upcasters: tt.fields.upcasters,
			}
			assert.Equalf(t, tt.want, c.Supports(tt.args.descriptor), "Supports(%v)", tt.args.descriptor)
		})
	}
}

func TestUpcasterChain_Upcast(t *testing.T) {
	type fields struct {
		upcasters []Upcaster
	}
	type args struct {
		descriptor UpcastableEventDescriptor
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []UpcastableEventDescriptor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := UpcasterChain{
				upcasters: tt.fields.upcasters,
			}
			assert.Equalf(t, tt.want, c.Upcast(tt.args.descriptor), "Upcast(%v)", tt.args.descriptor)
		})
	}
}

func TestUpcasterFunc_Supports(t *testing.T) {
	type args struct {
		descriptor UpcastableEventDescriptor
	}
	tests := []struct {
		name string
		u    UpcasterFunc
		args args
		want bool
	}{
		{
			name: "",
			u: func() (func(descriptor UpcastableEventDescriptor) bool, func(descriptor UpcastableEventDescriptor) []UpcastableEventDescriptor) {
				return func(descriptor UpcastableEventDescriptor) bool {
						return true
					}, func(descriptor UpcastableEventDescriptor) []UpcastableEventDescriptor {
						return []UpcastableEventDescriptor{descriptor}
					}
			},
			args: args{
				descriptor: UpcastableEventDescriptor{
					ID:             "",
					TypeName:       "",
					Payload:        nil,
					StreamID:       "",
					Version:        0,
					SequenceNumber: 0,
					RecordedAt:     time.Time{},
					Metadata:       nil,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.u.Supports(tt.args.descriptor), "Supports(%v)", tt.args.descriptor)
		})
	}
}

func TestUpcasterFunc_Upcast(t *testing.T) {
	type args struct {
		descriptor UpcastableEventDescriptor
	}
	tests := []struct {
		name string
		u    UpcasterFunc
		args args
		want []UpcastableEventDescriptor
	}{
		{
			name: "",
			u: func() (func(descriptor UpcastableEventDescriptor) bool, func(descriptor UpcastableEventDescriptor) []UpcastableEventDescriptor) {
				return func(descriptor UpcastableEventDescriptor) bool {
						return true
					}, func(descriptor UpcastableEventDescriptor) []UpcastableEventDescriptor {
						return []UpcastableEventDescriptor{descriptor}
					}
			},
			args: args{
				descriptor: UpcastableEventDescriptor{
					ID:             "#00",
					TypeName:       "unit.test",
					Payload:        nil,
					StreamID:       "",
					Version:        0,
					SequenceNumber: 0,
					RecordedAt:     time.Time{},
					Metadata:       nil,
				},
			},
			want: []UpcastableEventDescriptor{
				{
					ID:             "#00",
					TypeName:       "unit.test",
					Payload:        nil,
					StreamID:       "",
					Version:        0,
					SequenceNumber: 0,
					RecordedAt:     time.Time{},
					Metadata:       nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.u.Upcast(tt.args.descriptor), "Upcast(%v)", tt.args.descriptor)
		})
	}
}

func TestUpcastingEventStoreDecorator_ReadFromStream(t *testing.T) {

	streamID := StreamID("test")

	upcaster := UpcasterFunc(func() (func(descriptor UpcastableEventDescriptor) bool, func(descriptor UpcastableEventDescriptor) []UpcastableEventDescriptor) {
		return func(descriptor UpcastableEventDescriptor) bool {
				return descriptor.TypeName == "unit.test.upcastable"
			},
			func(descriptor UpcastableEventDescriptor) []UpcastableEventDescriptor {
				return []UpcastableEventDescriptor{descriptor.WithTypeName("unit.test.upcasted")}
			}
	})

	currentDate := time.Now()
	testClock := clock.NewFixedClock(currentDate)
	store := NewUpcastingEventStoreDecorator(NewInMemoryEventStore(testClock), NewUpcasterChain(upcaster))

	eventA := EventDescriptor{
		ID:       EventID(uuid.NewString()),
		TypeName: "unit.test.upcastable",
		Payload: EventPayload{
			"hello": "world",
		},
		Metadata: misas.Metadata{},
	}

	eventB := EventDescriptor{
		ID:       EventID(uuid.NewString()),
		TypeName: "unit.test.not-upcastable",
		Payload: EventPayload{
			"hello": "world",
		},
		Metadata: misas.Metadata{},
	}

	err := store.AppendToStream(context.Background(), streamID, []EventDescriptor{eventA, eventB}, WithOptimisticConcurrencyCheckDisabled())
	assert.NoError(t, err)

	stream, err := store.ReadFromStream(context.Background(), streamID, FromStart(), InForwardDirection())
	assert.NoError(t, err)

	assert.Equal(t, StreamSlice{
		StreamID: streamID,
		Descriptors: []RecordedEventDescriptor{
			{
				ID:       eventA.ID,
				StreamID: streamID,
				TypeName: "unit.test.upcasted",
				Payload: EventPayload{
					"hello": "world",
				},
				RecordedAt:     currentDate,
				Metadata:       misas.Metadata{},
				Version:        0,
				SequenceNumber: 0,
			},
			{
				ID:       eventB.ID,
				StreamID: streamID,
				TypeName: "unit.test.not-upcastable",
				Payload: EventPayload{
					"hello": "world",
				},
				RecordedAt:     currentDate,
				Metadata:       misas.Metadata{},
				Version:        1,
				SequenceNumber: 1,
			},
		},
	}, stream)
}
