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
	"github.com/morebec/misas-go/misas/event"
	"github.com/morebec/misas-go/misas/event/store"
	"github.com/pkg/errors"
)

// ProcessorOptions Represents a set of options that can be passed to an event.Processor to alter its behaviour.
type ProcessorOptions struct {
	Name                     string
	StreamID                 store.StreamID
	CheckpointCommitStrategy CheckpointCommitStrategy
	EventTypeNameFilter      *store.TypeNameFilter
}

type ProcessorOption func(options *ProcessorOptions)

// WithFiler Option allowing to filter events
func WithFiler(opts ...store.TypeNameFilterOption) ProcessorOption {
	return func(o *ProcessorOptions) {
		if len(opts) == 0 {
			o.EventTypeNameFilter = nil
		} else {
			for _, opt := range opts {
				opt(o.EventTypeNameFilter)
			}
		}
	}
}

// WithName Allows specifying the name of the Processor.
func WithName(name string) ProcessorOption {
	return func(options *ProcessorOptions) {
		options.Name = name
	}
}

// WithCheckpointCommitStrategy allows specifying the CheckpointCommitStrategy.
func WithCheckpointCommitStrategy(strategy CheckpointCommitStrategy) ProcessorOption {
	return func(options *ProcessorOptions) {
		options.CheckpointCommitStrategy = strategy
	}
}

// WithStreamId allows specifying which stream to process.
func WithStreamId(id store.StreamID) ProcessorOption {
	return func(options *ProcessorOptions) {
		options.StreamID = id
	}
}

// CheckpointCommitStrategy Represents the commit strategy to use for storing the checkpoints.
type CheckpointCommitStrategy string

// CommitBeforeProcessing Indicates that the checkpoint should be saved **before** processing the event.
// This strategy will mark the event as being processed, event in case of failure
// **Use this for at-most-once delivery.**
const CommitBeforeProcessing CheckpointCommitStrategy = "CommitBeforeProcessing"

// CommitAfterProcessing Indicates that the checkpoint should be saved **after** processing the event.
// In case of failure this will not mark the event as being processed.
// **Use this for at-least-once delivery.**
const CommitAfterProcessing CheckpointCommitStrategy = "CommitAfterProcessing"

type CheckpointID string

// Checkpoint represents a data structure that can be used to determine what was the last processed event in a stream.
type Checkpoint struct {
	ID       CheckpointID
	Position store.Position
	StreamID store.StreamID
}

// CheckpointStore allows storing checkpoints durably.
type CheckpointStore interface {
	// Save a checkpoint in this store, if the checkpoint already exists it is updated.
	Save(ctx context.Context, checkpoint Checkpoint) error

	// FindById returns a Checkpoint by its EventID or null if it was not found.
	FindById(ctx context.Context, id CheckpointID) (*Checkpoint, error)

	// Remove a checkpoint from this store. If the checkpoint does not exist, silently returns.
	Remove(ctx context.Context, id CheckpointID) error
}

type InMemoryCheckpointStore struct {
	checkpoints map[CheckpointID]Checkpoint
}

func NewInMemoryCheckpointStore() *InMemoryCheckpointStore {
	return &InMemoryCheckpointStore{checkpoints: map[CheckpointID]Checkpoint{}}
}

func (i InMemoryCheckpointStore) Save(_ context.Context, checkpoint Checkpoint) error {
	i.checkpoints[checkpoint.ID] = checkpoint
	return nil
}

func (i InMemoryCheckpointStore) FindById(_ context.Context, id CheckpointID) (*Checkpoint, error) {
	if c, ok := i.checkpoints[id]; !ok {
		return nil, errors.Errorf("checkpoint %s not found", id)
	} else {
		return &c, nil
	}
}

func (i InMemoryCheckpointStore) Remove(_ context.Context, id CheckpointID) error {
	delete(i.checkpoints, id)
	return nil
}

// Handler represents the type of work a Processor does with an event.
type Handler func(ctx context.Context, d store.RecordedEventDescriptor) error

// Processor is a service responsible for subscribing to a given stream of the event store in order to perform work with the events of the stream.
// It is intended to be run continuously.
// TODO tests
type Processor struct {
	eventStore      store.EventStore
	checkpointStore CheckpointStore
	options         ProcessorOptions
	running         bool
	processingFunc  Handler
}

// NewProcessor Creates a new Processor.
func NewProcessor(eventStore store.EventStore, checkpointStore CheckpointStore, processingFunc Handler, opts ...ProcessorOption) *Processor {
	if eventStore == nil {
		panic("cannot create a processor without event store")
	}

	if checkpointStore == nil {
		panic("cannot create a processor without checkpoint store")
	}

	if processingFunc == nil {
		panic("cannot create a processor without processing func")
	}

	options := ProcessorOptions{
		Name:                     "event-processor",
		StreamID:                 eventStore.GlobalStreamID(),
		CheckpointCommitStrategy: CommitAfterProcessing,
		EventTypeNameFilter:      nil,
	}

	for _, opt := range opts {
		opt(&options)
	}

	return &Processor{
		eventStore:      eventStore,
		checkpointStore: checkpointStore,
		options:         options,
		running:         false,
		processingFunc:  processingFunc,
	}
}

func (p *Processor) Run(ctx context.Context) (err error) {
	p.running = true

	defer func() {
		p.running = false
	}()

	// Subscribe to stream
	var filterOptions []store.TypeNameFilterOption
	if p.options.EventTypeNameFilter != nil {
		if p.options.EventTypeNameFilter.Mode == store.Exclude {
			filterOptions = append(filterOptions, store.ExcludeEventTypeNames(p.options.EventTypeNameFilter.EventTypeNames...))
		} else {
			filterOptions = append(filterOptions, store.SelectEventTypeNames(p.options.EventTypeNameFilter.EventTypeNames...))
		}
	}
	subscription, err := p.eventStore.SubscribeToStream(ctx, p.options.StreamID, store.WithSubscriptionFilter())

	if err != nil {
		return errors.Wrap(err, "failed processing events")
	}

	// Catchup
	if err := p.processEvents(ctx); err != nil {
		return errors.Wrap(err, "failed processing events")
	}

	// Listen for events
	for {
		select {
		case _ = <-subscription.EventChannel():
			if err := p.processEvents(ctx); err != nil {
				return errors.Wrap(err, "failed processing events")
			}
		case err := <-subscription.ErrorChannel():
			return errors.Wrap(err, "failed processing events")

		case <-ctx.Done():
			err := subscription.Close()
			return errors.Wrap(err, "failed processing events")
		}
	}
}

// Reset the stored checkpoint of this event processor. This can be used when a processor is required
// to be started anew from the beginning of the event store.Stream.
func (p *Processor) Reset(ctx context.Context) error {
	return p.checkpointStore.Remove(ctx, CheckpointID(p.options.Name))
}

func (p *Processor) processEvents(ctx context.Context) (err error) {
	// Get checkpoint
	checkpoint, err := p.fetchCheckpoint(ctx)
	if err != nil {
		return errors.Wrap(err, "failed updating event processor checkpoint")
	}

	stream, err := p.eventStore.ReadFromStream(ctx, p.options.StreamID, store.From(checkpoint.Position))
	if err != nil {
		return errors.Wrap(err, "failed updating event processor checkpoint")
	}

	for _, descriptor := range stream.Descriptors {
		// Update position
		checkpoint.Position = store.Position(descriptor.SequenceNumber)
		if p.options.CheckpointCommitStrategy == CommitBeforeProcessing {
			if err := p.checkpointStore.Save(ctx, checkpoint); err != nil {
				return errors.Wrap(err, "failed updating event processor checkpoint")
			}
		}

		if err := p.processingFunc(ctx, descriptor); err != nil {
			return errors.Wrapf(err, "failed processing event %s:%s", descriptor.TypeName, descriptor.ID)
		}

		if p.options.CheckpointCommitStrategy == CommitAfterProcessing {
			if err := p.checkpointStore.Save(ctx, checkpoint); err != nil {
				return errors.Wrap(err, "failed updating event processor checkpoint")
			}
		}
	}
	return nil
}

func (p *Processor) fetchCheckpoint(ctx context.Context) (Checkpoint, error) {
	if p.options.Name == "" {
		return Checkpoint{}, errors.New("cannot retrieve processor checkpoint: processor without a name")
	}
	checkpoint, _ := p.checkpointStore.FindById(ctx, CheckpointID(p.options.Name))
	//if err != nil {
	//	return Checkpoint{}, errors.Wrap(err, "failed retrieving event processor checkpoint")
	//}

	if checkpoint == nil {
		checkpoint = &Checkpoint{
			ID:       CheckpointID(p.options.Name),
			StreamID: p.options.StreamID,
			Position: store.Start,
		}
		err := p.checkpointStore.Save(ctx, *checkpoint)
		if err != nil {
			return Checkpoint{}, errors.Wrap(err, "failed initializing event processor checkpoint")
		}
	}

	return *checkpoint, nil
}

// SendToEventBusProcessingHandler  Returns a Processing Func that sends the store.RecordedEventDescriptor descriptor to an event.Bus.
func SendToEventBusProcessingHandler(eventConverter *store.EventConverter, bus event.Bus) Handler {
	return func(ctx context.Context, descriptor store.RecordedEventDescriptor) error {
		loaded, err := eventConverter.ConvertDescriptorToEvent(descriptor)
		if err != nil {
			return err
		}

		return bus.Send(ctx, loaded)
	}
}
