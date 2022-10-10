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
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func TestInMemoryCheckpointStore_FindById(t *testing.T) {
	type fields struct {
		checkpoints map[CheckpointID]Checkpoint
	}
	type args struct {
		ctx context.Context
		id  CheckpointID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Checkpoint
		wantErr bool
	}{
		{
			name: "test found",
			fields: fields{
				checkpoints: map[CheckpointID]Checkpoint{
					"00": {
						Position: store.Start,
						ID:       "00",
						StreamID: "$all",
					},
				},
			},
			args: args{
				ctx: context.Background(),
				id:  "00",
			},
			want: &Checkpoint{
				Position: store.Start,
				ID:       "00",
				StreamID: "$all",
			},
			wantErr: false,
		},
		{
			name: "test found",
			fields: fields{
				checkpoints: map[CheckpointID]Checkpoint{},
			},
			args: args{
				ctx: context.Background(),
				id:  "not-found",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := InMemoryCheckpointStore{
				checkpoints: tt.fields.checkpoints,
			}
			got, err := i.FindById(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindById() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInMemoryCheckpointStore_Remove(t *testing.T) {
	type fields struct {
		checkpoints map[CheckpointID]Checkpoint
	}
	type args struct {
		ctx context.Context
		id  CheckpointID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test remove",
			fields: fields{
				checkpoints: map[CheckpointID]Checkpoint{
					"00": {
						Position: store.Start,
						ID:       "00",
						StreamID: "$all",
					},
				},
			},
			args: args{
				ctx: context.Background(),
				id:  "00",
			},
			wantErr: false,
		},
		{
			name: "test remove not found should not cause errors",
			fields: fields{
				checkpoints: map[CheckpointID]Checkpoint{},
			},
			args: args{
				ctx: context.Background(),
				id:  "00",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := InMemoryCheckpointStore{
				checkpoints: tt.fields.checkpoints,
			}
			if err := i.Remove(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("Remove() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	s := InMemoryCheckpointStore{map[CheckpointID]Checkpoint{
		"00": {
			Position: store.Start,
			ID:       "00",
			StreamID: "$all",
		},
	}}
	_ = s.Remove(context.Background(), "00")
	// Ensure that it worked.
	found, err := s.FindById(context.Background(), "00")
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestInMemoryCheckpointStore_Save(t *testing.T) {
	type fields struct {
		checkpoints map[CheckpointID]Checkpoint
	}
	type args struct {
		ctx        context.Context
		checkpoint Checkpoint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test save",
			fields: fields{
				checkpoints: map[CheckpointID]Checkpoint{},
			},
			args: args{
				ctx: context.Background(),
				checkpoint: Checkpoint{
					ID:       "00",
					Position: store.Start,
					StreamID: "$all",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := InMemoryCheckpointStore{
				checkpoints: tt.fields.checkpoints,
			}
			if err := i.Save(tt.args.ctx, tt.args.checkpoint); (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	i := InMemoryCheckpointStore{map[CheckpointID]Checkpoint{}}
	err := i.Save(context.Background(), Checkpoint{
		ID:       "00",
		Position: 0,
		StreamID: "$all",
	})
	assert.NoError(t, err)
	found, err := i.FindById(context.Background(), "00")
	assert.NoError(t, err)
	assert.NotNil(t, found)
}

func TestNewInMemoryCheckpointStore(t *testing.T) {
	tests := []struct {
		name string
		want *InMemoryCheckpointStore
	}{
		{
			name: "test constructor",
			want: NewInMemoryCheckpointStore(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewInMemoryCheckpointStore(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewInMemoryCheckpointStore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_Reset(t *testing.T) {
	type fields struct {
		eventStore      store.EventStore
		checkpointStore CheckpointStore
		options         ProcessorOptions
		running         bool
		processingFunc  Handler
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
			name: "reset should not return errors",
			fields: fields{
				eventStore:      store.NewInMemoryEventStore(clock.NewUTCClock()),
				checkpointStore: InMemoryCheckpointStore{map[CheckpointID]Checkpoint{}},
				options:         ProcessorOptions{},
				running:         false,
				processingFunc:  nil,
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				eventStore:      tt.fields.eventStore,
				checkpointStore: tt.fields.checkpointStore,
				options:         tt.fields.options,
				running:         tt.fields.running,
				processingFunc:  tt.fields.processingFunc,
			}
			if err := p.Reset(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Reset() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProcessor_Run(t *testing.T) {
	type fields struct {
		eventStore      store.EventStore
		checkpointStore CheckpointStore
		options         ProcessorOptions
		running         bool
		processingFunc  Handler
	}
	type args struct {
		ctx context.Context
	}
	utcClock := clock.NewUTCClock()
	ctx, _ := context.WithDeadline(context.Background(), utcClock.Now().Add(5*time.Second))
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "reset should not return errors",
			fields: fields{
				eventStore:      store.NewInMemoryEventStore(utcClock),
				checkpointStore: InMemoryCheckpointStore{map[CheckpointID]Checkpoint{}},
				options: ProcessorOptions{
					Name:                     "test",
					StreamID:                 "$all",
					CheckpointCommitStrategy: CommitAfterProcessing,
					EventTypeNameFilter:      nil,
				},
				running:        false,
				processingFunc: nil,
			},
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Processor{
				eventStore:      tt.fields.eventStore,
				checkpointStore: tt.fields.checkpointStore,
				options:         tt.fields.options,
				running:         tt.fields.running,
				processingFunc:  tt.fields.processingFunc,
			}
			if err := p.Run(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
