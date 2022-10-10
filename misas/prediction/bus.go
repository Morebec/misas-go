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

package prediction

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
)

// Bus is a service responsible for abstracting the caller from a Handler.
type Bus interface {
	Send(ctx context.Context, p Prediction) error
	RegisterHandler(t TypeName, h Handler)
}

// InMemoryBus is an implementation of a prediction.Bus that sends predictions to handlers located in memory.
type InMemoryBus struct {
	handlers map[TypeName]Handler
}

func NewInMemoryBus() *InMemoryBus {
	return &InMemoryBus{map[TypeName]Handler{}}
}

// Send a prediction on this bus to be routed to its Prediction handler.
func (pb *InMemoryBus) Send(ctx context.Context, p Prediction) error {
	handler, err := pb.resolveHandler(p)
	if err != nil {
		// Prediction Handler should always be resolved! This is a critical error!
		panic(errors.Wrap(err, "failed sending prediction on bus"))
	}

	if err := handler.Handle(p, ctx); err != nil {
		return errors.Wrapf(err, "failed handling prediction \"%s\"", p.TypeName())
	}

	return nil
}

// RegisterHandler a handler for a given prediction type name.
func (pb *InMemoryBus) RegisterHandler(t TypeName, h Handler) {
	pb.handlers[t] = h
}

func (pb *InMemoryBus) resolveHandler(p Prediction) (Handler, error) {
	if handler, found := pb.handlers[p.TypeName()]; !found {
		return nil, errors.Errorf(fmt.Sprintf("no handler found for prediction \"%s\"", p.TypeName()))
	} else {
		return handler, nil
	}
}
