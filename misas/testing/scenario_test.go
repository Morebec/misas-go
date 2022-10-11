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
	"github.com/morebec/go-system/misas/clock"
	"github.com/morebec/go-system/misas/command"
	"github.com/morebec/go-system/misas/event"
	"github.com/morebec/go-system/misas/event/store"
	"github.com/morebec/go-system/misas/system"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type createAccount struct {
}

func (c createAccount) TypeName() command.TypeName {
	return "account.create"
}

type accountCreated struct {
}

func (a accountCreated) TypeName() event.TypeName {
	return "account.created"
}

func TestScenario(t *testing.T) {
	aClock := clock.NewFixedClock(time.Now())
	serv := system.New(
		system.WithSubsystems(
			func(m *system.Subsystem) {
				// m.RegisterEventHandler().Handles(accountCreated{})
				m.RegisterEvent(accountCreated{})
				m.RegisterCommandHandler(createAccount{}, command.HandlerFunc(func(ctx context.Context, c command.Command) (any, error) {
					payload, err := m.System.EventConverter.ToEventPayload(accountCreated{})
					if err != nil {
						return nil, err
					}

					if err = m.System.EventStore.AppendToStream(ctx, "test", []store.EventDescriptor{{
						ID:       store.NewEventID(),
						TypeName: accountCreated{}.TypeName(),
						Payload:  payload,
						Metadata: nil,
					}}); err != nil {
						return nil, err
					}

					return nil, nil
				}))
			},
		),
	)
	s := NewScenario(
		UsingService(serv),
		Given(
			CurrentDateIs(aClock.Now()),
		),
		When(
			Command(createAccount{}),
			// Query(),
			// Event(),
			// PredictionOccurs(),
		),
		Then(
			LastCommandBusErrorShouldBe(nil),
			LastCommandBusResponseShouldBe(nil),
			ExpectGlobalStream(
				HasRecorded(
					// NoEvents(),
					ExactlyTheseEvents(
						accountCreated{},
					),
				),
			),
		),
	)

	assert.NoError(t, s.Run(t))
}
