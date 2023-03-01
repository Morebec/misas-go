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
	"github.com/morebec/misas-go/misas/clock"
	"github.com/morebec/misas-go/misas/command"
	"github.com/morebec/misas-go/misas/event"
	"github.com/morebec/misas-go/misas/event/store"
	"github.com/morebec/misas-go/misas/system"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type createAccount struct {
}

func (c createAccount) TypeName() command.PayloadTypeName {
	return "account.create"
}

type accountCreated struct {
}

func (a accountCreated) TypeName() event.PayloadTypeName {
	return "account.created"
}

func TestScenario(t *testing.T) {
	aClock := clock.NewFixedClock(time.Now())
	sys := system.New(
		system.WithSubsystems(
			func(m *system.Subsystem) {
				// m.RegisterEventHandler().Handles(accountCreated{})
				m.RegisterEvent(accountCreated{})
				m.RegisterCommandHandler(createAccount{}.TypeName(), command.HandlerFunc(func(ctx context.Context, c command.Command) (any, error) {
					descriptor, err := m.System.EventConverter.ConvertEventToDescriptor(event.New(accountCreated{}))
					if err != nil {
						return nil, err
					}

					if err = m.System.EventStore.AppendToStream(ctx, "test", []store.EventDescriptor{descriptor}); err != nil {
						return nil, err
					}

					return nil, nil
				}))
			},
		),
	)
	s := NewScenario(
		UsingService(sys),
		Given(
			CurrentDateIs(aClock.Now()),
		),
		When(
			Command(command.Command{
				Payload: createAccount{},
			}),
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
						event.New(accountCreated{}),
					),
				),
			),
		),
	)

	assert.NoError(t, s.Run(t))
}
