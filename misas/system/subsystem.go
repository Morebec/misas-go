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

package system

import (
	"github.com/morebec/go-system/misas/command"
	"github.com/morebec/go-system/misas/event"
	"github.com/morebec/go-system/misas/prediction"
	"github.com/morebec/go-system/misas/query"
)

type Subsystem struct {
	System *System
}

// Environment returns the current environment of the System.
func (sub *Subsystem) Environment() Environment {
	return sub.System.Environment
}

// IsEnvironment indicated if the current environment is a given one.
func (sub *Subsystem) IsEnvironment(environment Environment) bool {
	return sub.Environment() == environment
}

// RegisterCommandHandler registers a command.Command and its command.Handler with the system.
func (sub *Subsystem) RegisterCommandHandler(c command.Command, h command.Handler) *Subsystem {
	sub.System.CommandBus.RegisterHandler(c.TypeName(), h)
	return sub
}

// RegisterQueryHandler registers a query.Query with the system.
func (sub *Subsystem) RegisterQueryHandler(q query.Query, h query.Handler) *Subsystem {
	sub.System.QueryBus.RegisterHandler(q.TypeName(), h)
	return sub
}

// RegisterEventHandler Register an event handler.
func (sub *Subsystem) RegisterEventHandler(h event.Handler) EventHandlerConfigurator {
	return EventHandlerConfigurator{handler: h, system: sub.System}
}

type EventHandlerConfigurator struct {
	system  *System
	handler event.Handler
}

// Handles allows registering the event and the handler with the event.Bus, as well as
// registering the event with the store.EventConverter.
func (c EventHandlerConfigurator) Handles(e event.Event) EventHandlerConfigurator {
	c.system.EventConverter.RegisterEvent(e)
	c.system.EventBus.RegisterHandler(e.TypeName(), c.handler)
	return c
}

// RegisterEvent allows registering an event with the store.EventConverter.
func (sub *Subsystem) RegisterEvent(e event.Event) *Subsystem {
	sub.System.EventConverter.RegisterEvent(e)
	return sub
}

// RegisterPredictionHandler registers a prediction.Handler.
func (sub *Subsystem) RegisterPredictionHandler(h prediction.Handler) PredictionHandlerConfigurator {
	return PredictionHandlerConfigurator{system: sub.System, handler: h}
}

type PredictionHandlerConfigurator struct {
	system  *System
	handler prediction.Handler
}

// Handles allows registering the prediction and the handler with the prediction.Bus, as well as
// registering the event with the prediction.Converter.
func (c PredictionHandlerConfigurator) Handles(p prediction.Prediction) PredictionHandlerConfigurator {
	c.system.PredictionConverter.RegisterPrediction(p)
	c.system.PredictionBus.RegisterHandler(p.TypeName(), c.handler)
	return c
}

// WithSubsystems specifies that the system uses a Subsystem.
func WithSubsystems(opts ...SubsystemConfigurator) Option {
	return func(s *System) {
		for _, opt := range opts {
			opt(&Subsystem{System: s})
		}
	}
}

type SubsystemConfigurator func(m *Subsystem)
