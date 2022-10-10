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

package processing

import (
	"context"
	"github.com/jwillp/go-system/misas/clock"
	"github.com/jwillp/go-system/misas/event/store"
	"github.com/pkg/errors"
	"testing"
	"time"
)

func TestProjectorGroup_Project(t *testing.T) {
	type fields struct {
		projectors []Projector
	}
	type args struct {
		ctx        context.Context
		descriptor store.RecordedEventDescriptor
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test no errors returned by inner projector does not cause errors for group.",
			fields: fields{
				projectors: []Projector{
					NewFuncProjector(func(ctx context.Context, descriptor store.RecordedEventDescriptor) error {
						return nil
					}, func(ctx context.Context) error {
						return nil
					}),
				},
			},
			args: args{
				ctx: context.Background(),
				descriptor: store.RecordedEventDescriptor{
					ID:             "",
					TypeName:       "",
					Payload:        nil,
					Metadata:       nil,
					StreamID:       "",
					Version:        0,
					SequenceNumber: 0,
					RecordedAt:     time.Time{},
				},
			},
			wantErr: false,
		},
		{
			name: "test error returned by inner projector causes error for group.",
			fields: fields{
				projectors: []Projector{
					NewFuncProjector(func(ctx context.Context, descriptor store.RecordedEventDescriptor) error {
						return errors.Errorf("There was an error.")
					}, func(ctx context.Context) error {
						return nil
					}),
				},
			},
			args: args{
				ctx: context.Background(),
				descriptor: store.RecordedEventDescriptor{
					ID:             "",
					TypeName:       "",
					Payload:        nil,
					Metadata:       nil,
					StreamID:       "",
					Version:        0,
					SequenceNumber: 0,
					RecordedAt:     time.Time{},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ProjectorGroup{
				projectors: tt.fields.projectors,
			}
			if err := p.Project(tt.args.ctx, tt.args.descriptor); (err != nil) != tt.wantErr {
				t.Errorf("Project() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProjectorGroup_Reset(t *testing.T) {
	type fields struct {
		projectors []Projector
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test no errors returned by inner projector does not cause errors for group.",
			fields: fields{
				projectors: []Projector{
					NewFuncProjector(func(ctx context.Context, descriptor store.RecordedEventDescriptor) error {
						return nil
					}, func(ctx context.Context) error {
						return nil
					}),
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
		{
			name: "test error returned by inner projector causes error for group.",
			fields: fields{
				projectors: []Projector{
					NewFuncProjector(func(ctx context.Context, descriptor store.RecordedEventDescriptor) error {
						return nil
					}, func(ctx context.Context) error {
						return errors.Errorf("There was an error.")
					}),
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ProjectorGroup{
				projectors: tt.fields.projectors,
			}
			if err := p.Reset(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Reset() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Example_projecting() {
	utcClock := clock.UTCClock{}
	eventStore := &store.InMemoryEventStore{Clock: utcClock}
	group := NewProjectorGroup()

	p := NewProcessor(eventStore, NewInMemoryCheckpointStore(), SendToProjectorProcessingHandler(group))
	err := p.Run(context.Background())
	if err != nil {
		panic(err)
	}
}
