package domain

import (
	"context"
	"github.com/morebec/misas-go/misas/clock"
	"github.com/morebec/misas-go/misas/command"
	"github.com/morebec/misas-go/misas/event"
	"github.com/morebec/misas-go/misas/event/store"
	"github.com/stretchr/testify/assert"
	"testing"
)

type registerUserCommand struct {
	ID           string
	EmailAddress string
}

func (r registerUserCommand) TypeName() command.TypeName {
	return "user.register"
}

type userRegisteredEvent struct {
	ID           string
	EmailAddress string
}

func (u userRegisteredEvent) TypeName() event.TypeName {
	return "user.registered"
}

type user struct {
	ID           string
	EmailAddress string
}

func userProjector() StateProjector[user] {
	return func(u user, e event.Event) user {
		switch e.(type) {
		case userRegisteredEvent:
			evt := e.(userRegisteredEvent)
			u.ID = evt.ID
			u.EmailAddress = evt.EmailAddress
		}
		return u
	}
}

func registerUser() StateHandler[user, registerUserCommand] {
	return func(u user, c registerUserCommand) (event.List, error) {
		return event.NewList(userRegisteredEvent{
			ID:           c.ID,
			EmailAddress: c.EmailAddress,
		}), nil
	}
}

func TestEventStreamCreatingCommandHandler(t *testing.T) {
	type args struct {
		repository       EventStoreRepository
		projector        StateProjector[user]
		handler          StateHandler[user, registerUserCommand]
		streamIdProvider StreamIDProviderFromState[user]
		responseProvider ResponseProvider[user]
		params           struct {
			context context.Context
			command command.Command
		}
	}
	tests := []struct {
		name string
		args args
		want struct {
			response any
			err      error
		}
	}{
		{
			name: "",
			args: args{
				repository: NewEventStoreRepository(
					store.NewInMemoryEventStore(clock.NewUTCClock()),
					store.NewEventConverter(),
					"user/",
					nil,
				),
				projector: userProjector(),
				handler:   registerUser(),
				streamIdProvider: func(u user) store.StreamID {
					return store.StreamID(u.ID)
				},
				responseProvider: func(u user) any {
					return u.ID
				},
				params: struct {
					context context.Context
					command command.Command
				}{
					context: context.Background(),
					command: registerUserCommand{
						ID:           "123",
						EmailAddress: "user@email.com",
					},
				},
			},
			want: struct {
				response any
				err      error
			}{response: "123", err: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := EventStreamCreatingCommandHandler(
				tt.args.repository,
				tt.args.projector,
				tt.args.handler,
				tt.args.streamIdProvider,
				tt.args.responseProvider,
			)
			response, err := handler(tt.args.params.context, tt.args.params.command)
			assert.Equal(t, tt.want.err, err)
			assert.Equalf(
				t,
				tt.want.response,
				response,
				"EventStreamCreatingCommandHandler(%v, %v, %v, %v, %v)",
				tt.args.repository,
				tt.args.projector,
				tt.args.handler,
				tt.args.streamIdProvider,
				tt.args.responseProvider,
			)
		})
	}
}

func TestVersion_Incremented(t *testing.T) {
	v := Version(5)
	assert.Equal(t, Version(6), v.Incremented())
}
