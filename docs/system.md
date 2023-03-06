# System
At the core of the library there is the concept of `System` which represents an information system.
The `System` struct is used as a centralized point to define systems, their subsystems and dependencies.
Although entirely optional, the use of the `System` struct allows to expressively define the dependencies
of the core units within the system.

## Defining a System
```go
	utcClock := clock.NewUTCClock()

	s := system.New(
		// These information are reused in logs, tracing spans or as metadata for events.
		system.WithInformation(system.Information{
			Name:    "unit_test",
			Version: "1.0.0",
		}),
		system.WithEnvironment(system.Test),
		system.WithClock(utcClock),

		system.WithCommandHandling(
			system.WithCommandBus(
				command.NewInMemoryBus(),
			),
		),

		system.WithQueryHandling(
			system.WithQueryBus(
				query.NewInMemoryBus(),
			),
		),

		system.WithEventHandling(
			system.WithEventBus(
				event.NewInMemoryBus(),
			),
			system.WithEventStore(
				postgresql.NewEventStore("connectionString", utcClock),
			),
		),

		system.WithPredictionHandling(
			system.WithPredictionBus(
				prediction.NewInMemoryBus(),
			),
			system.WithPredictionStore(
				postgresql.NewPredictionStore(),
			),
		),

		system.WithInstrumentation(
			system.WithTracer(instrumentation.NewSystemTracer()),
			system.WithDefaultLogger(),
			system.WithJaegerTracingSpanExporter("urlToJaeggerInstance"),
			system.WithCommandBusInstrumentation(), // Decorates the command bus adding automated instrumentation.
			system.WithQueryBusInstrumentation(), // Decorates the query bus adding automated instrumentation.
			system.WithEventBusInstrumentation(), // Decorates the event bus adding automated instrumentation.
			system.WithPredictionBusInstrumentation(), // Decorates the prediction bus adding automated instrumentation.
			system.WithEventStoreInstrumentation(), // Decorates the event store adding automated instrumentation.
		),

		// Modules allow separating the dependencies of the systems.
		system.WithSubsystems(
			func(s *system.Subsystem) {
				// Registers
				s.RegisterEvent(accountCreated{})
				s.RegisterCommandHandler(createAccount{}, createAccountCommandHandler))
			}, 
		),
	)
```
## Defining an Entry Point
Entry points are procedure to start the system and its subsystem's interaction layers.
Depending on the needs of the system, one could need to define different entry points
starting different things. (e.g. Web Server, Message Queue etc.)
```go
mainEntryPoint := NewEntryPoint(
		// Name of the entry point, if instrumentation is enabled (see below), this name will be used in spans.
		"web_server",

		// function that executions the entrypoint' main job, such as open 
		// a database connection, read configuration files etc.
		func(ctx context.Context, s *System) error {
			// Your logic here.
			return nil
		},
		
		// Allows adding automated instrumentation on the entry point.
		// Which are going to start and end spans with OTLP.
		WithEntryPointInstrumentation(),
	),
```

## Running an Entry Point
```go
// Allows running the system with the given entry point.
if err := s.RunEntryPoint(mainEntryPoint); err != nil {
    panic(err)
}
```

## Running Multiple Entry Points Concurrently
In most system it is necessary to define multiple entry points that should run concurrently.
For example, one entry point might be responsible for running the HTTP API Web Server, while another one
would be responsible for event processing or other background jobs.

The `RunAll` allows running these entry points concurrently, and returns a buffered channel of `ConcurrentRunResult`
that allows to be notified when an entrypoint terminates or fails:

```go
res := s.RunConcurrently(
		ctx,
		EventProcessorEntryPoint(s.EventStore, s.EventConverter, s.EventBus),
		FrontendApiServerEntryPoint(frontendApiServer),
		FrontendApiProjectorEntryPoint(frontendApiProjectorGroup, s.EventStore, s.EventConverter),
		TelegramIngesterEntryPoint(telegramIngester, channelRepository),
	)

	for r := range res {
		log.Printf("[%s] Terminated \n", r.EndpointName)
		if r.Err != nil {
			log.Printf("[%s] Failure: %s \n", r.EndpointName, r.Err.Error())
		}
	}
```


## Running All Entry Points of the System Concurrently
Although the `RunConcurrently` allows to specify exactly which endpoints to run, it can be simpler
to run all the endpoints that are registered with the system, using the `Run` method.