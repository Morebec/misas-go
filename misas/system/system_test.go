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
	"github.com/morebec/misas-go/misas/prediction"
	"github.com/morebec/misas-go/misas/query"
	"github.com/pkg/errors"
)

func ExampleSystem_Run() {
	// Defining a typed configuration struct
	type ExampleConfiguration struct {
		DbHost string
	}

	// This configuration can be loaded from a file or env variables.
	conf := ExampleConfiguration{
		DbHost: "host",
	}

	// Here's an example subsystem configuration that relies on the typed configuration.
	ExampleSubsystem := func(conf ExampleConfiguration) SubsystemConfigurator {
		return func(m *Subsystem) {

		}
	}

	utcClock := clock.NewUTCClock()
	s := New(
		WithInformation(Information{
			Name:    "unit_test",
			Version: "1.0.0",
		}),
		WithEnvironment(Test),
		WithClock(utcClock),

		WithCommandHandling(
			WithCommandBus(
				command.NewInMemoryBus(),
			),
		),

		WithQueryHandling(
			WithQueryBus(
				query.NewInMemoryBus(),
			),
		),

		WithEventHandling(
			WithEventBus(
				event.NewInMemoryBus(),
			),
			WithEventStore(
				store.NewInMemoryEventStore(utcClock),
			),
		),

		WithPredictionHandling(
			WithPredictionBus(
				prediction.NewInMemoryBus(),
			),
			WithPredictionStore(
				prediction.NewInMemoryStore(utcClock),
			),
		),

		WithInstrumentation(
			WithDefaultTracer(),
			WithDefaultLogger(),
			WithJaegerTracingSpanExporter(""),
			WithCommandBusInstrumentation(),
			WithQueryBusInstrumentation(),
			WithEventBusInstrumentation(),
			WithPredictionBusInstrumentation(),
			WithEventStoreInstrumentation(),
		),

		WithSubsystems(
			ExampleSubsystem(conf),
		),
	)

	mainEntryPoint := NewEntryPoint(
		"hello",
		func(ctx context.Context, s *System) error {
			// RunEntryPoint Web System, UpcastableEventPayload Processor, Background Process etc.
			return nil
		},
		WithEntryPointInstrumentation(),
	)

	err := s.RunEntryPoint(
		context.Background(),
		mainEntryPoint,
	)
	if err != nil {
		panic(errors.Wrap(err, "critical system error"))
	}
}
