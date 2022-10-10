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
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestIsStreamNotFoundError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "actual stream not found",
			args: args{
				err: NewStreamNotFoundError("unit"),
			},
			want: true,
		},
		{
			name: "actual stream not found",
			args: args{
				err: errors.New("was not a stream not found"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsStreamNotFoundError(tt.args.err), "IsStreamNotFoundError(%v)", tt.args.err)
		})
	}
}

func TestStreamSlice_First(t *testing.T) {
	type fields struct {
		StreamID    StreamID
		Descriptors []RecordedEventDescriptor
	}
	tests := []struct {
		name   string
		fields fields
		want   RecordedEventDescriptor
	}{
		{
			name: "first",
			fields: fields{
				StreamID: "unit",
				Descriptors: []RecordedEventDescriptor{
					{
						ID:             "0",
						TypeName:       "first",
						Payload:        nil,
						Metadata:       nil,
						StreamID:       "",
						Version:        0,
						SequenceNumber: 0,
						RecordedAt:     time.Time{},
					},
				},
			},
			want: RecordedEventDescriptor{
				ID:             "0",
				TypeName:       "first",
				Payload:        nil,
				Metadata:       nil,
				StreamID:       "",
				Version:        0,
				SequenceNumber: 0,
				RecordedAt:     time.Time{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := StreamSlice{
				StreamID:    tt.fields.StreamID,
				Descriptors: tt.fields.Descriptors,
			}
			assert.Equalf(t, tt.want, s.First(), "First()")
		})
	}
}

func TestStreamSlice_IsEmpty(t *testing.T) {
	type fields struct {
		StreamID    StreamID
		Descriptors []RecordedEventDescriptor
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "",
			fields: fields{
				StreamID:    "",
				Descriptors: nil,
			},
			want: true,
		},
		{
			name: "",
			fields: fields{
				StreamID: "",
				Descriptors: []RecordedEventDescriptor{
					{
						ID:             "01",
						TypeName:       "unit.test",
						Payload:        nil,
						Metadata:       nil,
						StreamID:       "",
						Version:        0,
						SequenceNumber: 0,
						RecordedAt:     time.Time{},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := StreamSlice{
				StreamID:    tt.fields.StreamID,
				Descriptors: tt.fields.Descriptors,
			}
			assert.Equalf(t, tt.want, s.IsEmpty(), "IsEmpty()")
		})
	}
}

func TestStreamSlice_Last(t *testing.T) {
	type fields struct {
		StreamID    StreamID
		Descriptors []RecordedEventDescriptor
	}
	tests := []struct {
		name   string
		fields fields
		want   RecordedEventDescriptor
	}{
		{
			name: "",
			fields: fields{
				StreamID: "",
				Descriptors: []RecordedEventDescriptor{
					{
						ID:             "00",
						TypeName:       "unit.test",
						Payload:        nil,
						Metadata:       nil,
						StreamID:       "",
						Version:        0,
						SequenceNumber: 0,
						RecordedAt:     time.Time{},
					},
					{
						ID:             "01",
						TypeName:       "unit.test",
						Payload:        nil,
						Metadata:       nil,
						StreamID:       "",
						Version:        0,
						SequenceNumber: 0,
						RecordedAt:     time.Time{},
					},
				},
			},
			want: RecordedEventDescriptor{
				ID:             "01",
				TypeName:       "unit.test",
				Payload:        nil,
				Metadata:       nil,
				StreamID:       "",
				Version:        0,
				SequenceNumber: 0,
				RecordedAt:     time.Time{},
			},
		},
		{
			name: "",
			fields: fields{
				StreamID:    "",
				Descriptors: nil,
			},
			want: RecordedEventDescriptor{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := StreamSlice{
				StreamID:    tt.fields.StreamID,
				Descriptors: tt.fields.Descriptors,
			}
			assert.Equalf(t, tt.want, s.Last(), "Last()")
		})
	}
}

func TestStreamSlice_Length(t *testing.T) {
	type fields struct {
		StreamID    StreamID
		Descriptors []RecordedEventDescriptor
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "",
			fields: fields{
				StreamID:    "",
				Descriptors: nil,
			},
			want: 0,
		},
		{
			name: "",
			fields: fields{
				StreamID: "",
				Descriptors: []RecordedEventDescriptor{
					{
						ID:             "00",
						TypeName:       "unit.test",
						Payload:        nil,
						Metadata:       nil,
						StreamID:       "",
						Version:        0,
						SequenceNumber: 0,
						RecordedAt:     time.Time{},
					},
					{
						ID:             "01",
						TypeName:       "unit.test",
						Payload:        nil,
						Metadata:       nil,
						StreamID:       "",
						Version:        0,
						SequenceNumber: 0,
						RecordedAt:     time.Time{},
					},
				},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := StreamSlice{
				StreamID:    tt.fields.StreamID,
				Descriptors: tt.fields.Descriptors,
			}
			assert.Equalf(t, tt.want, s.Length(), "Length()")
		})
	}
}

func TestStreamSlice_Reversed(t *testing.T) {
	type fields struct {
		StreamID    StreamID
		Descriptors []RecordedEventDescriptor
	}
	tests := []struct {
		name   string
		fields fields
		want   StreamSlice
	}{
		{
			name: "",
			fields: fields{
				StreamID: "unit.test",
				Descriptors: []RecordedEventDescriptor{
					{
						ID:             "00",
						TypeName:       "unit.test",
						Payload:        nil,
						Metadata:       nil,
						StreamID:       "",
						Version:        0,
						SequenceNumber: 0,
						RecordedAt:     time.Time{},
					},
					{
						ID:             "01",
						TypeName:       "unit.test",
						Payload:        nil,
						Metadata:       nil,
						StreamID:       "",
						Version:        0,
						SequenceNumber: 0,
						RecordedAt:     time.Time{},
					},
				},
			},
			want: StreamSlice{
				StreamID: "unit.test",
				Descriptors: []RecordedEventDescriptor{
					{
						ID:             "01",
						TypeName:       "unit.test",
						Payload:        nil,
						Metadata:       nil,
						StreamID:       "",
						Version:        0,
						SequenceNumber: 0,
						RecordedAt:     time.Time{},
					},
					{
						ID:             "00",
						TypeName:       "unit.test",
						Payload:        nil,
						Metadata:       nil,
						StreamID:       "",
						Version:        0,
						SequenceNumber: 0,
						RecordedAt:     time.Time{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := StreamSlice{
				StreamID:    tt.fields.StreamID,
				Descriptors: tt.fields.Descriptors,
			}
			assert.Equalf(t, tt.want, s.Reversed(), "Reversed()")
		})
	}
}

func TestStreamSlice_Select(t *testing.T) {
	type fields struct {
		StreamID    StreamID
		Descriptors []RecordedEventDescriptor
	}
	type args struct {
		p func(event RecordedEventDescriptor) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []RecordedEventDescriptor
	}{
		{
			name: "",
			fields: fields{
				StreamID: "unit.test",
				Descriptors: []RecordedEventDescriptor{
					{
						ID:             "00",
						TypeName:       "unit.test",
						Payload:        nil,
						Metadata:       nil,
						StreamID:       "",
						Version:        0,
						SequenceNumber: 0,
						RecordedAt:     time.Time{},
					},
					{
						ID:             "01",
						TypeName:       "unit.test",
						Payload:        nil,
						Metadata:       nil,
						StreamID:       "",
						Version:        0,
						SequenceNumber: 0,
						RecordedAt:     time.Time{},
					},
				},
			},
			args: args{
				p: func(event RecordedEventDescriptor) bool {
					return event.ID == "00"
				},
			},
			want: []RecordedEventDescriptor{
				{
					ID:             "00",
					TypeName:       "unit.test",
					Payload:        nil,
					Metadata:       nil,
					StreamID:       "",
					Version:        0,
					SequenceNumber: 0,
					RecordedAt:     time.Time{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := StreamSlice{
				StreamID:    tt.fields.StreamID,
				Descriptors: tt.fields.Descriptors,
			}
			assert.Equalf(t, tt.want, s.Select(tt.args.p), "Select(%v)", tt.args.p)
		})
	}
}
