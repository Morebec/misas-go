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
	"github.com/morebec/go-system/misas/event"
	"github.com/morebec/go-system/misas/event/store"
)

// A Projector is a service responsible for projecting events to Read Models.
// This read model can be stored in any storage deemed fit for the given read model.
type Projector interface {
	// Reset Allows resting this projector to its initial state.
	// Essentially it should bring back all its read models to their initial state or delete them entirely.
	// This is used when a projection needs to be entirely recomputed (replayed).
	Reset(ctx context.Context) error

	// Project an event to a read model.
	Project(ctx context.Context, descriptor store.RecordedEventDescriptor) error
}

// SendToProjectorProcessingHandler returns a processing handler that sends an event to a given projector.
func SendToProjectorProcessingHandler(p Projector) Handler {
	return func(ctx context.Context, d store.RecordedEventDescriptor) error {
		return p.Project(ctx, d)
	}
}

// ProjectorGroup implementation of a Projector that delegates te work of projecting to other projectors.
type ProjectorGroup struct {
	projectors []Projector
}

func NewProjectorGroup(projectors ...Projector) *ProjectorGroup {
	return &ProjectorGroup{projectors: projectors}
}

func (p ProjectorGroup) Reset(ctx context.Context) error {
	for _, i := range p.projectors {
		if err := i.Reset(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (p ProjectorGroup) Project(ctx context.Context, descriptor store.RecordedEventDescriptor) error {
	for _, i := range p.projectors {
		if err := i.Project(ctx, descriptor); err != nil {
			return err
		}
	}

	return nil
}

// FuncProjector implementation of a Projector that allows defining a Projector using functions.
type FuncProjector struct {
	ProjectFunc func(ctx context.Context, descriptor store.RecordedEventDescriptor) error
	ResetFunc   func(ctx context.Context) error
}

func NewFuncProjector(projectFunc func(ctx context.Context, descriptor store.RecordedEventDescriptor) error, resetFunc func(ctx context.Context) error) *FuncProjector {
	return &FuncProjector{ProjectFunc: projectFunc, ResetFunc: resetFunc}
}

func (p FuncProjector) Reset(ctx context.Context) error {
	return p.ResetFunc(ctx)
}

func (p FuncProjector) Project(ctx context.Context, descriptor store.RecordedEventDescriptor) error {
	return p.ProjectFunc(ctx, descriptor)
}

type ConvertedEventProjector interface {
	// Reset Allows resting this projector to its initial state.
	// Essentially it should bring back all its read models to their initial state or delete them entirely.
	// This is used when a projection needs to be entirely recomputed (replayed).
	Reset(ctx context.Context) error

	// Project an event to a read model.
	Project(ctx context.Context, e event.Event, descriptor store.RecordedEventDescriptor) error
}

// StreamToConvertedEventProjector returns a function that allows calling a ConvertedEventProjector
func StreamToConvertedEventProjector(eventStore store.EventStore, eventConverter store.EventConverter, projector ConvertedEventProjector) func(ctx context.Context, streamId store.StreamID) error {
	return func(ctx context.Context, streamId store.StreamID) error {
		stream, err := eventStore.ReadFromStream(ctx, streamId)
		if err != nil {
			return err
		}
		for _, d := range stream.Descriptors {
			e, err := eventConverter.FromRecordedEventDescriptor(d)
			if err != nil {
				return err
			}
			if err := projector.Project(ctx, e, d); err != nil {
				return err
			}
		}
		return nil
	}
}
