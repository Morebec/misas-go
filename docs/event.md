# Messaging: Events

## Implement a Event
```go
type UserRegisteredEvent struct {
	EmailAddress string `json:"emailAddress"`
	Password     string `json:"password"`
}

func (c ImplementEvent) TypeName() event.PayloadTypeName {
	return "user.register"
}
```

## Implement a Event Handler
```go
func UserRegisteredEventHandler() event.HandlerFunc {
	return func(ctx context.Context, evt event.Event) (any, error) {
	    evt, ok := evt.Payload.(UserRegisteredEvent)
	    if !ok {
                return errors.New("invalid_event")
            }
        // Implement logic ...
    }
}
```

## Register an Event Handler with the Event Bus
```go
cb := event.NewInMemoryBus()
cb.RegisterHandler(UserRegisteredEvent{}.TypeName(), UserRegisteredEventHandler())
```

## Send event to the event bus
```go
cb := event.NewInMemoryBus()
cb.Send(context.Background(), event.New(UserRegisteredEvent{
	Username: "misas",
	Password: "a_password"
}))
```
