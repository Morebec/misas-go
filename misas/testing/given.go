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
	"fmt"
	"github.com/morebec/misas-go/misas/clock"
	"github.com/morebec/misas-go/misas/event"
	"github.com/morebec/misas-go/misas/event/store"
	"github.com/stretchr/testify/assert"
	"time"
)

type GivenOption func(scenario *Scenario, stage *Stage)

func Given(opts ...GivenOption) ScenarioOption {
	return func(s *Scenario) {
		stage := &Stage{
			Name:  "Given",
			Steps: []Step{},
		}
		for _, o := range opts {
			o(s, stage)
		}
		s.addStage(*stage)
	}
}

// CurrentDateIs allows specifying a step where the date of the clock will be set to the provided value.
// Note that the clock must be a clock.FixedClock for this to take effect. otherwise it will return an error.
func CurrentDateIs(dt time.Time) GivenOption {
	return func(scenario *Scenario, s *Stage) {
		s.addStep(NewStep("setCurrentClockDateTime", func(t assert.TestingT, scenario *Scenario, s *Stage) error {
			serviceClock := scenario.Clock()
			c, _ := serviceClock.(clock.FixedClock)
			//if !ok {
			//	return errors.New("Can only set the current date on a clock.FixedClock")
			//}
			c.CurrentDate = dt
			return nil
		}))
	}
}

type EventStreamOption func(id store.StreamID, scenario *Scenario, stage *Stage)

// EventStream allows configuring a certain EventStream Given Stage.
func EventStream(id store.StreamID, opts ...EventStreamOption) GivenOption {
	return func(scenario *Scenario, s *Stage) {
		for _, opt := range opts {
			opt(id, scenario, s)
		}
	}
}

// RecordedEvent allows specifying that a certain EventStream has recorded a given event as part of a Given Stage.
func RecordedEvent(event event.Event, opts ...store.AppendToStreamOption) EventStreamOption {
	return func(id store.StreamID, scenario *Scenario, stage *Stage) {
		stage.addStep(NewStep(fmt.Sprintf("recordEventToStream->%s", id), func(t assert.TestingT, scenario *Scenario, s *Stage) error {

			payload, err := scenario.Service.EventConverter.ConvertEventToDescriptor(event)
			if err != nil {
				return err
			}

			descriptor := store.EventDescriptor{
				ID:       store.NewEventID(),
				TypeName: event.Payload.TypeName(),
				Payload:  payload,
				Metadata: nil,
			}

			return scenario.EventStore().AppendToStream(scenario.Execution.Context, id, []store.EventDescriptor{descriptor}, opts...)
		}))
	}
}
