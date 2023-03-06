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

package system

import (
	"context"
	"github.com/morebec/misas-go/misas/clock"
	"github.com/morebec/misas-go/misas/command"
	"github.com/morebec/misas-go/misas/event"
	"github.com/morebec/misas-go/misas/event/store"
	"github.com/morebec/misas-go/misas/instrumentation"
	"github.com/morebec/misas-go/misas/prediction"
	"github.com/morebec/misas-go/misas/query"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel/sdk/trace"
	"sync"
)

type Environment string

const (
	Dev     Environment = "dev"
	Test    Environment = "Test"
	Staging Environment = "staging"
	Prod    Environment = "prod"
)

type Information struct {
	Name    string
	Version string
}

// System is a simple struct representing the core dependencies of a System.
// It should only be used to set up the dependencies of the System's services, in the entry points of the System.
// SERVICES SHOULD NOT RELY ON THE System INSTANCE AS A DEPENDENCY TO AVOID ANY GOD OBJECT-LIKE PATTERN.
type System struct {
	Environment Environment
	Information Information

	Clock clock.Clock

	CommandBus command.Bus
	QueryBus   query.Bus

	EventBus           event.Bus
	EventStore         store.EventStore
	EventConverter     *store.EventConverter
	EventUpcasterChain *store.UpcasterChain

	PredictionBus           prediction.Bus
	PredictionStore         prediction.Store
	PredictionConverter     *prediction.Converter
	PredictionUpcasterChain *prediction.UpcasterChain

	Logger       *otelzap.Logger
	Tracer       *instrumentation.SystemTracer
	SpanExporter trace.SpanExporter
	Services     Services
	EntryPoints  []EntryPoint
}

type Option func(*System)

func WithInformation(information Information) Option {
	return func(system *System) {
		system.Information = information
	}
}

func WithEnvironment(environment Environment) Option {
	return func(system *System) {
		system.Environment = environment
	}
}

// New Creates a System Instance with sane defaults.
func New(opts ...Option) *System {
	systemClock := clock.UTCClock{}

	system := &System{
		Environment: Dev,
		Information: Information{
			Name:    "unknown System",
			Version: "0.0.1",
		},
		Clock:               systemClock,
		CommandBus:          command.NewInMemoryBus(),
		QueryBus:            query.NewInMemoryBus(),
		EventBus:            event.NewInMemoryBus(),
		EventStore:          store.NewInMemoryEventStore(systemClock),
		EventConverter:      store.NewEventConverter(),
		PredictionBus:       prediction.NewInMemoryBus(),
		PredictionStore:     prediction.NewInMemoryStore(systemClock),
		PredictionConverter: prediction.NewConverter(),
		Logger:              nil,
		Tracer:              nil,
		Services:            Services{},
		EntryPoints:         nil,
	}

	for _, opt := range opts {
		opt(system)
	}

	return system
}

// RunEntryPoint Allows running a EntryPoint with the system, passing down a context.
// if the entry point returns an error it will be returned by this function.
// Entry points should handle errors as they see fit.
func (s *System) RunEntryPoint(ctx context.Context, entry EntryPoint) error {
	return entry.Run(ctx, s)
}

// ConcurrentRunResult is a data structure representing the concurrent execution of an endpoint with the System.RunConcurrently method.
type ConcurrentRunResult struct {
	EndpointName string
	Err          error
}

// RunConcurrently allows running multiple EntryPoint concurrently each inside their own goroutines.
// This method returns a channel that allows listening being notified when an endpoint terminates.
func (s *System) RunConcurrently(ctx context.Context, entryPoints ...EntryPoint) chan ConcurrentRunResult {

	wg := sync.WaitGroup{}
	runChan := make(chan ConcurrentRunResult, len(entryPoints))

	for _, entryPoint := range entryPoints {
		wg.Add(1)
		e := entryPoint
		go func() {
			defer wg.Done()
			runResult := ConcurrentRunResult{EndpointName: e.Name()}
			if err := s.RunEntryPoint(ctx, e); err != nil {
				runResult.Err = err
			}
			runChan <- runResult
		}()
	}

	go func() {
		wg.Wait()
		close(runChan)
	}()

	return runChan
}

// Run allows running all Entry points of the system.
func (s *System) Run(ctx context.Context) chan ConcurrentRunResult {
	return s.RunConcurrently(ctx, s.EntryPoints...)
}

// Service returns a service by its name or nil if it was not found.
func (s *System) Service(name string) Service {
	if serv, ok := s.Services[name]; !ok {
		return nil
	} else {
		return serv
	}
}

// EntryPoint returns an entry point by its name or nil if not found.
func (s *System) EntryPoint(name string) EntryPoint {
	for _, e := range s.EntryPoints {
		if e.Name() == name {
			return e
		}
	}

	return nil
}
