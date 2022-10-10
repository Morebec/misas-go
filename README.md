# MISAS Go

MISAS Go is an opinionated library to easily develop systems using a predefined architecture using DDD, CQRS and ES.
It provides a solid base for smaller teams to develop advanced system with lesser means.
It is the go implementation of [MISAS](https://github.com/Morebec/misas).

## Features
- Domain-Driven Design
- Event Sourcing
- CQRS
- Intra/Out of Process Messaging
- Observability
  - Tracing using Open Telemetry as automated instrumentation on Command/Query/Event/Prediction buses. 
  - Tracing using correlation ID and causation ID on messages propagated to events.

## Introduction
MISAS Go mostly provides a set of abstractions to implement DDD, CQRS and ES concepts according to [MISAS](https://github.com/Morebec/misas).
It also provides a few concrete implementation of these concepts, for the most common use cases.


### Defining a System
At the core of the library there is the concept of `System` which represents an information system.
The `System` struct is used as a centralized point to define systems.
Although entirely optional, the use of the `System` allows to expressively define the dependencies
of the core units withing the system: 


```go
	utcClock := clock.NewUTCClock()

	s := system.New(
		// These information are reused in logs, tracing spans or as metadata for events.
		WithInformation(Information{
			Name:    "unit_test",
			Version: "1.0.0",
		}),
		WithEnvironment(system.Test),
		WithClock(utcClock),

		WithCommandHandling(
			WithCommandBus(
				command.NewInMemoryBus(),
			),
		),

		WithQueryHandling(
			WithQueryBus(
				query.NewInMemoryBus(),
			),
		),

		WithEventHandling(
			WithEventBus(
				event.NewInMemoryBus(),
			),
			WithEventStore(
				postgresql.NewEventStore(utcClock),
			),
		),

		WithPredictionHandling(
			WithPredictionBus(
				prediction.NewInMemoryBus(),
			),
			WithPredictionStore(
				postgresql.NewPredictionStore(),
			),
		),

		WithInstrumentation(
			WithTracer(instrumentation.NewSystemTracer()),
			WithDefaultLogger(),
			WithJaegerTracingSpanExporter("urlToJaeggerInstance"),
			WithCommandBusInstrumentation(), // Decorates the command bus adding automated instrumentation.
			WithQueryBusInstrumentation(), // Decorates the query bus adding automated instrumentation.
			WithEventBusInstrumentation(), // Decorates the event bus adding automated instrumentation.
			WithPredictionBusInstrumentation(), // Decorates the prediction bus adding automated instrumentation.
			WithEventStoreInstrumentation(), // Decorates the event store adding automated instrumentation.
		),

		// Modules allow separating the dependencies of the systems.
		WithModules(
			func(m *system.Module) {
				// Registers
				m.RegisterEvent(accountCreated{})
				m.RegisterCommand(createAccount{}, createAccountCommandHandler))
			}, 
		),
	)

	// Entry points are procedure to start the system and its interaction layer.
	// Depending on the needs of the system, one could need to define different entry points
	// starting different things. (e.g. Web Server, Message Queue etc.)
    mainEntryPoint := NewEntryPoint(
		// Name of the entry point, if instrumentation is enabled (see below), this name will be used in spans.
		"web_server",

		// Function to effectively start the entry point.
		func(ctx context.Context, s *System) error {
			return nil
		},

		// Function to stop the entry point.
		func(ctx context.Context, s *System) error {
			return nil
		},

		// Allows adding automated instrumentation on the entry point.
		WithEntryPointInstrumentation(),
	),
	
	// Allows running the system with the given entry point.
	if err := s.Run(mainEntryPoint); err != nil {
		panic(err)
	}
```

## Command Processing

### Command Handlers & Failures

### Registering a Command Handler with the System

## Aggregates

### Implementing an Aggregate
The aggregate interface has the following structure:
```go
type Aggregate interface {
	// Apply an event to this aggregate without recording it as an uncommitted change.
	Apply(e event.Event)

	// RecordAndApplyEvent Records an event.Event as an uncommitted event and applies it to this aggregate.
	RecordAndApplyEvent(e event.Event)

	// UncommittedEvents Returns the list of uncommitted events on this aggregate.
	UncommittedEvents() event.List

	// ClearUncommittedEvents Clears the list of uncommitted events on this aggregate.
	ClearUncommittedEvents()
}
```

To simplify the work of the implementors of this interface, the embeddable `EventRecorder` struct can be used.
This result in the implementor only requiring to implement the `Apply(e event.Event)` and `RecordAndApplyEvent(e event.Event)` methods.

```go
type User struct {
	EventRecorder
	ID           string
	EmailAddress string
}

func (u *User) Apply(e event.Event) {
	switch e.(type) {
	case UserRegisteredEvent:
		evt := e.(UserRegisteredEvent)
		u.ID = evt.ID
		u.EmailAddress = evt.EmailAddress
	}
}

func (u *User) RecordAndApplyEvent(e event.Event) error {
	u.RecordEvent(e) // method provided by the EventRecorder
	u.Apply(e)
	return nil
}
```

> Note: The method `RecordAndApplyEvent(e event.Event)` can be a useful place to perform general safe guarding for the end of life of an aggregate by returning an error
> in case of violated invariants.
> 
> *For example, when a user account is banned, it should not be possible to perform any other changes to the account.*

### Aggregate Repositories
You can use the aggregate.EventStoreRepository helper to quickly implement repositories for your aggregates through composition:

```go
type UserRepository struct {
  inner: aggregate.EventStoreRepository
}

func NewUserRepository(es event.Store) *UserRepository {
  return &UserRepository{
    inner: aggregate.NewEventStoreRepository(es, func() aggregate.Aggregate { 
	  // This callback allows defining the initial state of an aggregate before applying its saved changes
          // when loading.
      return &User{}
    }),
  }
}

func (r *UserRepository) Add(ctx context.Context, u *User) error {
    return r.inner.Add(ctx, event.StreamID("user/"+u.ID), u)	
}

func (r *UserRepository) Save(ctx context.Context, u *User, version Version) error {
    return r.inner.Save(ctx, event.StreamID("user/"+u.ID), u, version)	
}

func (r *UserRepository) FindByID(ctx context.Context, id UserID) (*User, Version, error) {
  loaded, v, err := r.inner.Load(ctx, event.StreamID("user/"+u.ID))
  if err != nil {
    return &User{}, 0, err
  }
  
  return loaded.(*User), v, nil
}
```


## Query Processing

### Query Handlers & Failures

## Event Processing


### Registering a Query Handler with the System

### Event Store

### Event Store Subscriptions

### Event Processors

## Event Handlers
```go
eventBus.RegisterHandler(EventTypeName, Handler)
```

### Checkpoints


### Event Handlers & Failures
When an event handler fails to process an event, there are two common strategies at our disposal:
- **Continued Processing:** Ignore/log the failure and continue processing the next events.
- **Delayed Processing:** Stop/retry the processing at the problematic event, until fixed.

Each strategy has its own pros and cons. 
**Delayed Processing** prevents any out-of-order processing of events, and ensures that the system when done with the processing will be fully consistent.
However, it will require the 
event handlers to be idempotent since they have the potential of being called multiple times for the same events in cases of retries.
**Continued Processing** has the benefit of not blocking the processing of events and can therefore minimize the impact it has on other components of the system,
however, it also means that event handlers should be implemented in a way to support inconsistencies in data since some events will have happened and will hve been partially
applied. This leads to a system that can be slightly inconsistent, and will require close attention to these potential inconsistencies.

An interesting strategy is to used Delayed processing combined with event processing partitions, 
(e.g. one event processor per module) which can often drastically minimize the bottlenecks occasioned by having a problematic event. 

## Registering Event Handlers with the System

## Prediction Processing

### Prediction Handlers & Failures


## Projection Processing

### Projectors & Failures