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
	"github.com/morebec/go-system/misas/command"
	"github.com/morebec/go-system/misas/event"
	"github.com/morebec/go-system/misas/prediction"
	"github.com/morebec/go-system/misas/query"
	"github.com/stretchr/testify/assert"
)

type WhenOption func(scenario *Scenario, stage *Stage)

func When(opts ...WhenOption) ScenarioOption {
	return func(s *Scenario) {
		stage := &Stage{
			Name:  "When",
			Steps: []Step{},
		}
		for _, o := range opts {
			o(s, stage)
		}
		s.addStage(*stage)
	}
}

// Command Allows adding a step to send a Command to the command.Bus
func Command(c command.Command) WhenOption {
	return func(scenario *Scenario, stage *Stage) {
		stage.addStep(NewStep("sendCommandToCommandBus", func(t assert.TestingT, scenario *Scenario, stage *Stage) error {
			response, err := scenario.CommandBus().Send(scenario.Execution.Context, c)
			scenario.Execution.LastCommandBusResponse = response
			scenario.Execution.LastCommandBusError = err
			return nil
		}))
	}
}

// Query Allows adding a step to send a Query to the query.Bus
func Query(q query.Query) WhenOption {
	return func(scenario *Scenario, stage *Stage) {
		stage.addStep(NewStep("sendQueryToQueryBus", func(t assert.TestingT, scenario *Scenario, stage *Stage) error {
			response, err := scenario.QueryBus().Send(scenario.Execution.Context, q)
			scenario.Execution.LastQueryBusResponse = response
			scenario.Execution.LastQueryBusError = err
			return nil
		}))
	}
}

// Event Allows adding a step to send an Event to the event.Bus
func Event(e event.Event) WhenOption {
	return func(scenario *Scenario, stage *Stage) {
		stage.addStep(NewStep("sendEventToEventBus", func(t assert.TestingT, scenario *Scenario, stage *Stage) error {
			err := scenario.EventBus().Send(scenario.Execution.Context, e)
			scenario.Execution.LastEventBusError = err
			return nil
		}))
	}
}

// PredictionOccurs Allows adding a step to send a prediction.Prediction to the prediction.Bus
func PredictionOccurs(p prediction.Prediction) WhenOption {
	return func(scenario *Scenario, stage *Stage) {
		stage.addStep(NewStep("sendPredictionToPredictionBus", func(t assert.TestingT, scenario *Scenario, stage *Stage) error {
			err := scenario.PredictionBus().Send(scenario.Execution.Context, p)
			scenario.Execution.LastPredictionBusError = err
			return nil
		}))
	}
}
