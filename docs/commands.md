# Messaging: Commands

## Implement a Command
```go
type RegisterUserCommand struct {
	EmailAddress string `json:"emailAddress"`
	Password     string `json:"password"`
}

func (c ImplementCommand) TypeName() command.PayloadTypeName {
	return "user.register"
}
```

## Implement a Command Handler
```go
func RegisterUserCommandHandler() command.HandlerFunc {
	return func(ctx context.Context, cmd command.Command) (any, error) {
	    c, ok := cmd.Payload.(RegisterUserCommand)
	    if !ok {
                return errors.New("invalid_command")
            }
        // Implement logic ...
    }
}
```

## Register a Command Handler with the Command Bus
```go
cb := command.NewInMemoryBus()
cb.RegisterHandler(RegisterUserCommand{}.TypeName(), RegisterUserCommandHandler())
```

## Send command to the command bus
```go
cb := command.NewInMemoryBus()
cb.Send(context.Background(), command.New(RegisterUserCommand{
	Username: "misas",
	Password: "a_password"
}))
```