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

package testing

import (
	"context"
	"github.com/jwillp/go-system/misas/clock"
	"github.com/jwillp/go-system/misas/command"
	"github.com/jwillp/go-system/misas/event"
	"github.com/jwillp/go-system/misas/event/store"
	"github.com/jwillp/go-system/misas/prediction"
	"github.com/jwillp/go-system/misas/query"
	"github.com/jwillp/go-system/misas/system"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"time"
)

// Scenario represents a test scenario. A Test scenario is made out of Stage (Given, When Then).
// These stages are further divided up into Step.
type Scenario struct {
	stages    []Stage
	Service   *system.System
	Execution *ScenarioExecution
}

type ScenarioExecution struct {
	Context                context.Context
	Scenario               *Scenario
	CurrentStage           Stage
	CurrentStep            Step
	LastCommandBusResponse any
	LastCommandBusError    error
	LastQueryBusResponse   any
	LastQueryBusError      error
	LastEventBusError      error
	LastPredictionBusError error

	RecordedEvents         []store.RecordedEventDescriptor
	lastSequenceNumberRead store.SequenceNumber
}

func (e *ScenarioExecution) Run(t assert.TestingT) error {
	if err := e.initializeEventTracking(); err != nil {
		return errors.Wrap(err, "failed running scenario")
	}

	for _, stage := range e.Scenario.stages {
		if err := e.runStage(t, stage); err != nil {
			return errors.Wrap(err, "failed running scenario")
		}
	}

	return nil
}

func (e *ScenarioExecution) runStage(t assert.TestingT, stage Stage) error {
	e.CurrentStage = stage
	for _, step := range stage.Steps {
		if err := e.runStep(t, stage, step); err != nil {
			return errors.Wrapf(err, "failed running stage \"%s\"", stage.Name)
		}
	}
	return nil
}

func (e *ScenarioExecution) runStep(t assert.TestingT, stage Stage, step Step) error {
	e.CurrentStep = step
	if err := step.Run(t, e.Scenario, &stage); err != nil {
		return errors.Wrapf(err, "failed running step \"%s\"", step.Name)
	}

	// Track events
	if err := e.trackEvents(); err != nil {
		return errors.Wrapf(err, "failed tracking events after step \"%s\"", step.Name)
	}
	return nil
}

func (e *ScenarioExecution) initializeEventTracking() error {
	eventStore := e.Scenario.EventStore()
	e.lastSequenceNumberRead = store.SequenceNumber(store.Start)

	stream, err := eventStore.ReadFromStream(e.Context, eventStore.GlobalStreamID(), store.LastEvent())
	if err != nil {
		return errors.Wrap(err, "failed initializing event tracking")
	}

	if stream.IsEmpty() {
		return nil
	}

	e.lastSequenceNumberRead = stream.Last().SequenceNumber
	return nil
}

func (e *ScenarioExecution) trackEvents() error {
	eventStore := e.Scenario.EventStore()
	stream, err := eventStore.ReadFromStream(
		e.Context,
		eventStore.GlobalStreamID(),
		store.InForwardDirection(),
		store.From(store.Position(e.lastSequenceNumberRead)),
	)

	if err != nil {
		return errors.Wrap(err, "failed tracking events")
	}

	if stream.IsEmpty() {
		return nil
	}

	for _, descriptor := range stream.Descriptors {
		e.RecordedEvents = append(e.RecordedEvents, descriptor)
	}

	e.lastSequenceNumberRead = stream.Last().SequenceNumber

	return nil
}

func (e *ScenarioExecution) RecordedEventsByStreamId(id store.StreamID) []store.RecordedEventDescriptor {
	eventStore := e.Scenario.EventStore()
	if id == eventStore.GlobalStreamID() {
		return e.RecordedEvents
	}

	var events []store.RecordedEventDescriptor
	for _, evt := range e.RecordedEvents {
		if evt.StreamID == id {
			events = append(events, evt)
		}
	}

	return events
}

type ScenarioOption func(s *Scenario)

func NewScenario(options ...ScenarioOption) *Scenario {
	s := &Scenario{
		stages: []Stage{},
		Service: system.New(
			system.WithEnvironment(system.Test),
			system.WithClock(clock.NewFixedClock(time.Now())),
		),
	}

	for _, opt := range options {
		opt(s)
	}

	return s
}

// AddStage to the scenario.
func (s *Scenario) addStage(st Stage) {
	s.stages = append(s.stages, st)
}

// Run the scenario.
func (s *Scenario) Run(t assert.TestingT) error {
	exec := &ScenarioExecution{
		Context:      context.Background(),
		Scenario:     s,
		CurrentStage: Stage{},
		CurrentStep:  Step{},
	}
	s.Execution = exec
	return exec.Run(t)
}

func (s *Scenario) Clock() clock.Clock {
	return s.Service.Clock
}

func (s *Scenario) EventStore() store.EventStore {
	return s.Service.EventStore
}

func (s *Scenario) CommandBus() command.Bus {
	return s.Service.CommandBus
}

func (s *Scenario) QueryBus() query.Bus {
	return s.Service.QueryBus
}

func (s *Scenario) EventBus() event.Bus {
	return s.Service.EventBus
}

func (s *Scenario) PredictionBus() prediction.Bus {
	return s.Service.PredictionBus
}

func UsingService(s *system.System) ScenarioOption {
	return func(sc *Scenario) {
		sc.Service = s
	}
}

func WithClock(c clock.Clock) ScenarioOption {
	return func(s *Scenario) {
		s.Service.Clock = c
	}
}

func WithEventStore(e store.EventStore) ScenarioOption {
	return func(s *Scenario) {
		s.Service.EventStore = e
	}
}

func WithEnvironment(e system.Environment) ScenarioOption {
	return func(s *Scenario) {
		s.Service.Environment = e
	}
}

// Stage represents a stage in a test. A Test Scenario is made of multiple stages that each perform specific tasks known as Step.
// in simpler terms, stages correspond to the Given, When and Then of a scenario.
type Stage struct {
	Name  string
	Steps []Step
}

// addStep to this Stage.
func (s *Stage) addStep(step Step) {
	s.Steps = append(s.Steps, step)
}

// Step in a Stage.
type Step struct {
	Name string
	Run  StepFunction
}

func NewStep(name string, run StepFunction) Step {
	return Step{Name: name, Run: run}
}

// StepFunction represents a function to run a given step.
type StepFunction func(t assert.TestingT, scenario *Scenario, stage *Stage) error
