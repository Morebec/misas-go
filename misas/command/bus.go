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

package command

import (
	"context"
	"github.com/pkg/errors"
)

// Bus is a service responsible for decoupling a caller and the handler of a command.
type Bus interface {

	// RegisterHandler Registers a new Handler for a given Command with the bus.
	// If a handler was already registered for a given command, it should be replaced by the new one provided.
	RegisterHandler(t PayloadTypeName, h Handler)

	// Send a Command to its Handler for processing and returns a potential payload acting as a fulfillment result.
	// The fulfillment result might be the identity of the resource/concept introduced in the system, an identifier
	// representing a location/read model where the result of the command's fulfilment can be obtained.
	// if no handler could be found for the given command, an error should be returned.
	Send(ctx context.Context, c Command) (any, error)
}

// InMemoryBus is an implementation of a command.Bus that sends command to handlers in memory.
// The processing of the commands by their handlers is performed synchronously in the same memory space.
type InMemoryBus struct {
	handlers map[PayloadTypeName]Handler
}

// NewInMemoryBus allows constructing an InMemoryBus.
func NewInMemoryBus() *InMemoryBus {
	bus := &InMemoryBus{
		handlers: map[PayloadTypeName]Handler{},
	}
	return bus
}

func (cb *InMemoryBus) RegisterHandler(t PayloadTypeName, h Handler) {
	cb.handlers[t] = h
}

func (cb *InMemoryBus) Send(ctx context.Context, c Command) (any, error) {
	// Handle
	events, err := cb.handleCommand(ctx, c)
	if err != nil {
		return nil, err
	}

	return events, nil
}

// handles a command by finding its resolver.
func (cb *InMemoryBus) handleCommand(ctx context.Context, c Command) (any, error) {

	handler, err := cb.resolveHandler(c.Payload.TypeName())
	if err != nil {
		err = errors.Wrapf(err, "failed handling command \"%s\"", c.Payload.TypeName())
		// Command should always be resolved! This is a critical error!
		panic(err)
	}

	events, err := handler.Handle(ctx, c)
	if err != nil {
		err = errors.Wrapf(err, "failed handling command \"%s\"", c.Payload.TypeName())
		return nil, err
	}
	return events, nil
}

// resolves the handler associated with a certain command type.
func (cb *InMemoryBus) resolveHandler(ptn PayloadTypeName) (Handler, error) {
	if handler, found := cb.handlers[ptn]; !found {
		return nil, errors.Errorf("no handler found for command \"%s\"", ptn)
	} else {
		return handler, nil
	}
}
