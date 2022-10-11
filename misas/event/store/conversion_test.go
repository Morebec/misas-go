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
	"fmt"
	"github.com/morebec/go-system/misas/event"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

type eventLoaded struct {
	AString string
	AnInt   int
	AFloat  float64
	ABool   bool
	ARune   rune
	AMap    map[string]any
	AList   []string
}

const eventLoadedTypeName event.TypeName = "event.loaded"

func (l eventLoaded) TypeName() event.TypeName {
	return eventLoadedTypeName
}

func TestEventLoader_Load(t *testing.T) {
	type fields struct {
		events map[event.TypeName]reflect.Type
	}
	type args struct {
		descriptor RecordedEventDescriptor
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    event.Event
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "",
			fields: fields{
				events: map[event.TypeName]reflect.Type{},
			},
			args: args{
				descriptor: RecordedEventDescriptor{
					ID:       "#000",
					TypeName: eventLoadedTypeName,
					Payload: EventPayload{
						"AString": "string",
						"AnInt":   1,
						"AFloat":  50.25,
						"ABool":   true,
						"ARune":   'A',
						"AMap": map[string]any{
							"hello": "world",
						},
						"AList": []string{
							"hello",
							"world",
						},
					},
					Metadata:       nil,
					StreamID:       "unit.test",
					Version:        0,
					SequenceNumber: 0,
					RecordedAt:     time.Time{},
				},
			},
			want: eventLoaded{
				AString: "string",
				AnInt:   1,
				AFloat:  50.25,
				ABool:   true,
				ARune:   'A',
				AMap: map[string]any{
					"hello": "world",
				},
				AList: []string{
					"hello",
					"world",
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &EventConverter{
				events: tt.fields.events,
			}
			e.RegisterEvent(eventLoaded{})
			got, err := e.FromRecordedEventDescriptor(tt.args.descriptor)
			if !tt.wantErr(t, err, fmt.Sprintf("FromRecordedEventDescriptor(%v)", tt.args.descriptor)) {
				return
			}
			assert.Equalf(t, tt.want, got, "FromRecordedEventDescriptor(%v)", tt.args.descriptor)
		})
	}
}

func TestEventConverter_ToEventPayload(t *testing.T) {
	type fields struct {
		events map[event.TypeName]reflect.Type
	}
	type args struct {
		evt event.Event
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    EventPayload
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "",
			fields: fields{
				events: nil,
			},
			args: args{
				evt: eventLoaded{
					AString: "string",
					AnInt:   50,
					AFloat:  50.25,
					ABool:   true,
					ARune:   'A',
					AMap:    map[string]any{"hello": "world"},
					AList:   []string{"hello", "world"},
				},
			},
			want: EventPayload{
				"AString": "string",
				"AnInt":   50.0,  // marshalling as json causes numbers which are translated to float64
				"AFloat":  50.25, // marshalling as json causes numbers which are translated to float64
				"ABool":   true,
				"ARune":   65.0, // marshalling as json causes numbers which are translated to float64
				"AMap": map[string]any{
					"hello": "world",
				},
				"AList": []any{
					"hello",
					"world",
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &EventConverter{
				events: tt.fields.events,
			}
			got, err := e.ToEventPayload(tt.args.evt)
			if !tt.wantErr(t, err, fmt.Sprintf("ToEventPayload(%v)", tt.args.evt)) {
				return
			}
			assert.Equalf(t, tt.want, got, "ToEventPayload(%v)", tt.args.evt)
		})
	}
}
