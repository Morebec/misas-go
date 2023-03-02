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
	"log"
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
	}

	for _, opt := range opts {
		opt(system)
	}

	return system
}

// Run Allows running the System with a managed context.
func (s *System) Run(ctx context.Context, entry EntryPoint) error {
	var err error
	go func() {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		err = entry.Start(ctx, s)
	}()

	_ = <-ctx.Done()
	if err := entry.Stop(ctx, s); err != nil {
		return err
	}

	return err
}

func (s *System) RunAll(ctx context.Context, entryPoints ...EntryPoint) error {

	var reterr error
	wg := sync.WaitGroup{}

	for _, e := range entryPoints {
		wg.Add(1)
		e := e
		go func() {
			defer wg.Done()
			if err := s.Run(ctx, e); err != nil {
				log.Print(err.Error())
				reterr = err
			}
		}()
	}

	wg.Wait()

	return reterr
}

// Service returns a service by its name or nil if it was not found.
func (s *System) Service(name string) Service {
	if serv, ok := s.Services[name]; !ok {
		return nil
	} else {
		return serv
	}
}
