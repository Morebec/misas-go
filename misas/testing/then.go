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
	"github.com/morebec/misas-go/misas/event"
	"github.com/morebec/misas-go/misas/event/store"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type ThenOption func(scenario *Scenario, stage *Stage)

func Then(opts ...ThenOption) ScenarioOption {
	return func(s *Scenario) {
		stage := &Stage{
			Name:  "Then",
			Steps: []Step{},
		}
		for _, o := range opts {
			o(s, stage)
		}
		s.addStage(*stage)
	}
}

type ExpectEventStreamOption func(id store.StreamID, scenario Scenario, stage *Stage)

// ExpectEventStream allows specifying expectations about a certain event stream.
func ExpectEventStream(id store.StreamID, opts ...ExpectEventStreamOption) ThenOption {
	return func(scenario *Scenario, stage *Stage) {
		for _, opt := range opts {
			opt(id, *scenario, stage)
		}
	}
}

// ExpectGlobalStream allows specifying expectations about the event store's global stream.
func ExpectGlobalStream(opts ...ExpectEventStreamOption) ThenOption {
	return func(scenario *Scenario, stage *Stage) {
		id := scenario.EventStore().GlobalStreamID()
		for _, opt := range opts {
			opt(id, *scenario, stage)
		}
	}
}

type HasRecordedOption func(id store.StreamID, scenario Scenario, stage *Stage)

// HasRecorded allows specifying expectations about a certain EventStream having recorded certain events.
func HasRecorded(opts ...HasRecordedOption) ExpectEventStreamOption {
	return func(id store.StreamID, scenario Scenario, stage *Stage) {
		for _, opt := range opts {
			opt(id, scenario, stage)
		}
	}
}

// ExactlyOneEvent allows specifying the expectation that a given stream has recorded only a single event as provided.
func ExactlyOneEvent(expectedEvent event.Event) HasRecordedOption {
	return func(id store.StreamID, scenario Scenario, stage *Stage) {
		stage.addStep(NewStep("ExpectOnlyOneEvent", func(t assert.TestingT, scenario *Scenario, stage *Stage) error {
			actualEvents := scenario.Execution.RecordedEventsByStreamId(id)

			if !assert.Len(t, actualEvents, 1) {
				return errors.Errorf("Failed asserting that event only one event was recorded in stream %s, got %d", id, len(actualEvents))
			}

			actualEvent, err := scenario.Service.EventConverter.FromRecordedEventDescriptor(actualEvents[0])
			if err != nil {
				return err
			}
			assert.Equal(t, expectedEvent, actualEvent)

			return nil
		}))
	}
}

// NoEvents allows specifying the expectation that a given stream has recorded no events at all.
func NoEvents() HasRecordedOption {
	return func(id store.StreamID, scenario Scenario, stage *Stage) {
		stage.addStep(NewStep("ExpectNoEvents", func(t assert.TestingT, scenario *Scenario, stage *Stage) error {

			actualEvents := scenario.Execution.RecordedEventsByStreamId(id)
			if !assert.Empty(t, actualEvents) {
				return errors.Errorf("failed asserting there were no events in stream %s, got %d events", id, len(actualEvents))
			}
			return nil
		}))
	}
}

// ExactlyTheseEvents allows specifying the expectation that a given stream has recorded exactly the events as provided in order.
func ExactlyTheseEvents(expectedEvents ...event.Event) HasRecordedOption {
	return func(id store.StreamID, scenario Scenario, stage *Stage) {
		stage.addStep(NewStep("ExpectTheseEvents", func(t assert.TestingT, scenario *Scenario, stage *Stage) error {
			actualEvents := scenario.Execution.RecordedEventsByStreamId(id)
			if !assert.Equal(t, len(expectedEvents), len(actualEvents)) {
				return fmt.Errorf("%d events recorded, when %d expected in stream %s", len(expectedEvents), len(actualEvents), id)
			}

			for _, expectedEvent := range expectedEvents {
				for _, actualEventDescriptor := range actualEvents {
					actualEvent, err := scenario.Service.EventConverter.FromRecordedEventDescriptor(actualEventDescriptor)
					if err != nil {
						return err
					}
					assert.Equal(t, expectedEvent, actualEvent)
				}
			}

			return nil
		}))
	}
}

// LastCommandBusResponseShouldBe allows specifying an expectation for the last command bus response.
func LastCommandBusResponseShouldBe(expectedResponse any) ThenOption {
	return func(scenario *Scenario, stage *Stage) {
		stage.addStep(NewStep("LastCommandBusResponseShouldBe", func(t assert.TestingT, scenario *Scenario, stage *Stage) error {
			actualResponse := scenario.Execution.LastCommandBusResponse
			if !assert.Equal(t, expectedResponse, actualResponse) {
				lastCommandBusError := scenario.Execution.LastCommandBusError
				if lastCommandBusError != nil {
					return errors.Errorf("failed asserting that the last command bus response was %v, got an error instead: %v", expectedResponse, lastCommandBusError)
				} else {
					return errors.Errorf("failed asserting that the last command bus response was %v, got %v", expectedResponse, actualResponse)
				}
			}
			return nil
		}))
	}
}

// LastCommandBusResponseShouldBeString allows specifying an expectation for the last command bus response to be a string.
func LastCommandBusResponseShouldBeString() ThenOption {
	return LastCommandBusResponseShould(func(t assert.TestingT, scenario Scenario, actualResponse any) error {
		if !assert.IsType(t, "", actualResponse) {
			return errors.Errorf("failed asserting that the last command bus response was a string, got %v", actualResponse)
		}
		return nil
	})
}

// LastCommandBusErrorShouldBe  allows specifying an expectation for the command bus to have responses with an error
func LastCommandBusErrorShouldBe(expectedError error) ThenOption {
	return func(scenario *Scenario, stage *Stage) {
		stage.addStep(NewStep("LastCommandBusErrorShouldBe", func(t assert.TestingT, scenario *Scenario, stage *Stage) error {
			actualError := scenario.Execution.LastCommandBusError
			if !assert.Equal(t, expectedError, actualError) {
				return errors.Errorf("failed asserting that the last command bus error was %v, got %v", expectedError, actualError)
			}
			return nil
		}))
	}
}

// LastCommandBusResponseShould allows specifying an expectation using a func to allow advanced use cases the last command bus response.
func LastCommandBusResponseShould(expectation func(t assert.TestingT, scenario Scenario, actualResponse any) error) ThenOption {
	return func(scenario *Scenario, stage *Stage) {
		stage.addStep(NewStep("LastCommandBusResponseShould", func(t assert.TestingT, scenario *Scenario, stage *Stage) error {
			return expectation(t, *scenario, scenario.Execution.LastCommandBusResponse)
		}))
	}
}

// LastCommandBusErrorShould allows specifying an expectation using a func to allow advanced use cases the last command bus error.
func LastCommandBusErrorShould(expectation func(t assert.TestingT, scenario Scenario, actualError any) error) ThenOption {
	return func(scenario *Scenario, stage *Stage) {
		stage.addStep(NewStep("LastCommandBusErrorShould", func(t assert.TestingT, scenario *Scenario, stage *Stage) error {
			return expectation(t, *scenario, scenario.Execution.LastCommandBusError)
		}))
	}
}

// LastQueryBusResponseShouldBe allows specifying an expectation for the last query bus response.
func LastQueryBusResponseShouldBe(expectedResponse any) ThenOption {
	return func(scenario *Scenario, stage *Stage) {
		stage.addStep(NewStep("LastQueryBusResponseShouldBe", func(t assert.TestingT, scenario *Scenario, stage *Stage) error {
			actualResponse := scenario.Execution.LastQueryBusResponse
			if !assert.Equal(t, expectedResponse, actualResponse) {
				lastQueryBusError := scenario.Execution.LastQueryBusError
				if lastQueryBusError != nil {
					return errors.Errorf("failed asserting that the last query bus response was %v, got an error instead: %v", expectedResponse, lastQueryBusError)
				} else {
					return errors.Errorf("failed asserting that the last query bus response was %v, got: %v", expectedResponse, actualResponse)
				}
			}
			return nil
		}))
	}
}

// LastQueryBusResponseShould allows specifying an expectation using a func to allow advanced use cases for the last query bus response.
func LastQueryBusResponseShould(expectation func(t assert.TestingT, scenario Scenario, actualResponse any) error) ThenOption {
	return func(scenario *Scenario, stage *Stage) {
		stage.addStep(NewStep("LastQueryBusResponseShould", func(t assert.TestingT, scenario *Scenario, stage *Stage) error {
			return expectation(t, *scenario, scenario.Execution.LastQueryBusResponse)
		}))
	}
}

// LastQueryBusErrorShould allows specifying an expectation using a func to allow advanced use cases the last query bus error.
func LastQueryBusErrorShould(expectation func(t assert.TestingT, scenario Scenario, actualError any) error) ThenOption {
	return func(scenario *Scenario, stage *Stage) {
		stage.addStep(NewStep("LastQueryBusErrorShould", func(t assert.TestingT, scenario *Scenario, stage *Stage) error {
			return expectation(t, *scenario, scenario.Execution.LastQueryBusError)
		}))
	}
}

// LastQueryBusErrorShouldBe  allows specifying an expectation for the query bus to have responses with an error
func LastQueryBusErrorShouldBe(expectedError error) ThenOption {
	return func(scenario *Scenario, stage *Stage) {
		stage.addStep(NewStep("LastQueryBusErrorShouldBe", func(t assert.TestingT, scenario *Scenario, stage *Stage) error {
			actualError := scenario.Execution.LastQueryBusError
			if !assert.Equal(t, expectedError, actualError) {
				return errors.Errorf("failed asserting that the last query bus error was %v, got: %v", expectedError, actualError)
			}
			return nil
		}))
	}
}

// LastEventBusErrorShould allows specifying an expectation using a func to allow advanced use cases the last event bus error.
func LastEventBusErrorShould(expectation func(t assert.TestingT, scenario Scenario, actualError any) error) ThenOption {
	return func(scenario *Scenario, stage *Stage) {
		stage.addStep(NewStep("LastEventBusErrorShould", func(t assert.TestingT, scenario *Scenario, stage *Stage) error {
			return expectation(t, *scenario, scenario.Execution.LastCommandBusError)
		}))
	}
}

// LastEventBusErrorShouldBe  allows specifying an expectation for the event bus to have responses with an error
func LastEventBusErrorShouldBe(expectedError error) ThenOption {
	return func(scenario *Scenario, stage *Stage) {
		stage.addStep(NewStep("LastEventBusErrorShouldBe", func(t assert.TestingT, scenario *Scenario, stage *Stage) error {
			actualError := scenario.Execution.LastEventBusError
			if !assert.Equal(t, expectedError, actualError) {
				return errors.Errorf("failed asserting that the last event bus error was %v, got: %v", expectedError, actualError)
			}
			return nil
		}))
	}
}

// LastPredictionBusErrorShould allows specifying an expectation using a func to allow advanced use cases the last prediction bus error.
func LastPredictionBusErrorShould(expectation func(t assert.TestingT, scenario Scenario, actualError any) error) ThenOption {
	return func(scenario *Scenario, stage *Stage) {
		stage.addStep(NewStep("LastPredictionBusErrorShould", func(t assert.TestingT, scenario *Scenario, stage *Stage) error {
			return expectation(t, *scenario, scenario.Execution.LastPredictionBusError)
		}))
	}
}

// LastPredictionBusErrorShouldBe  allows specifying an expectation for the prediction bus to have responses with an error
func LastPredictionBusErrorShouldBe(expectedError error) ThenOption {
	return func(scenario *Scenario, stage *Stage) {
		stage.addStep(NewStep("LastPredictionBusErrorShouldBe", func(t assert.TestingT, scenario *Scenario, stage *Stage) error {
			actualError := scenario.Execution.LastPredictionBusError
			if !assert.Equal(t, expectedError, actualError) {
				return errors.Errorf("failed asserting that the last prediction bus error was %v, got: %v", expectedError, actualError)
			}
			return nil
		}))
	}
}
