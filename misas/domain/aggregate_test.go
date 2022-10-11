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

package domain

import (
	"context"
	"fmt"
	"github.com/morebec/misas-go/misas/clock"
	"github.com/morebec/misas-go/misas/event"
	"github.com/morebec/misas-go/misas/event/store"
	"github.com/stretchr/testify/assert"
	"testing"
)

type UserRegisteredEvent struct {
	ID           string
	EmailAddress string
}

func (u UserRegisteredEvent) TypeName() event.TypeName {
	return "user.registered"
}

type User struct {
	EventSourcedAggregateBase
	ID           string
	EmailAddress string
}

func (u *User) ApplyEvent(evt event.Event) {
	switch evt.(type) {
	case UserRegisteredEvent:
		e := evt.(UserRegisteredEvent)
		u.ID = e.ID
		u.EmailAddress = e.EmailAddress
	}
}

func RegisterUser(id string, emailAddress string) *User {
	u := &User{
		ID:           "",
		EmailAddress: "",
	}
	u.EventSourcedAggregateBase = EventSourcedAggregateBase{
		ApplyEvent: u.ApplyEvent,
	}

	u.RecordEvent(UserRegisteredEvent{
		ID:           id,
		EmailAddress: emailAddress,
	})

	return u
}

func TestEventStoreRepository_Add(t *testing.T) {
	es := store.NewInMemoryEventStore(clock.UTCClock{})

	evts := RegisterUser("0", "john@email.com")

	repo := NewEventStoreRepository[*User](es, store.NewEventConverter())
	err := repo.Add(context.Background(), "user/0", evts)
	assert.Nil(t, err)
}

func TestEventStoreRepository_Update(t *testing.T) {

}

func TestEventStoreRepository_Load(t *testing.T) {
	es := store.NewInMemoryEventStore(clock.UTCClock{})
	converter := store.NewEventConverter()
	converter.RegisterEvent(UserRegisteredEvent{})
	repo := NewEventStoreRepository[*User](es, converter)

	streamID := store.StreamID("user/usr_123")

	load, v, err := repo.Load(context.Background(), streamID)
	if err != nil {
		return
	}
	fmt.Println(load)
	fmt.Println(v)

	err = repo.Add(context.Background(), streamID, RegisterUser("user_123", "user@email.com"))
	assert.Nil(t, err)

	loaded, version, err := repo.Load(context.Background(), streamID)
	assert.Nil(t, err)
	assert.Equal(t, Version(0), version)

	assert.Equal(t, User{
		ID:           "user_123",
		EmailAddress: "user@email.com",
	}, loaded)
}

func TestVersion_Incremented(t *testing.T) {
	v := Version(5)
	assert.Equal(t, Version(6), v.Incremented())
}
