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
	"github.com/morebec/misas-go/misas/event"
	"github.com/stretchr/testify/assert"
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

const eventLoadedTypeName event.PayloadTypeName = "event.loaded"

func (l eventLoaded) TypeName() event.PayloadTypeName {
	return eventLoadedTypeName
}

func TestEventLoader_ConvertDescriptorToEvent(t *testing.T) {
	type args struct {
		descriptor RecordedEventDescriptor
	}
	tests := []struct {
		name    string
		args    args
		want    event.Event
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "default",
			args: args{
				descriptor: RecordedEventDescriptor{
					ID:       "#000",
					TypeName: eventLoadedTypeName,
					Payload: DescriptorPayload{
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
			want: event.New(eventLoaded{
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
			}),
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEventConverter()
			e.RegisterEventPayload(eventLoaded{})
			got, err := e.ConvertDescriptorToEvent(tt.args.descriptor)
			if !tt.wantErr(t, err, fmt.Sprintf("ConvertDescriptorToEvent(%v)", tt.args.descriptor)) {
				return
			}
			assert.Equalf(t, tt.want, got, "ConvertDescriptorToEvent(%v)", tt.args.descriptor)
		})
	}
}

func TestEventConverter_ConvertEventToDescriptor(t *testing.T) {
	type args struct {
		evt event.Event
	}
	tests := []struct {
		name    string
		args    args
		want    EventDescriptor
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "default",
			args: args{
				evt: event.New(eventLoaded{
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
				}),
			},
			want: EventDescriptor{
				ID:       "",
				TypeName: "",
				Payload:  nil,
				Metadata: nil,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewEventConverter()
			got, err := c.ConvertEventToDescriptor(tt.args.evt)
			if !tt.wantErr(t, err, fmt.Sprintf("ConvertEventToDescriptor(%v)", tt.args.evt)) {
				return
			}
			assert.Equalf(t, tt.want, got, "ConvertEventToDescriptor(%v)", tt.args.evt)
		})
	}
}
