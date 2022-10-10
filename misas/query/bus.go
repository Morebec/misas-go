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

package query

import (
	"context"
	"github.com/pkg/errors"
)

// Bus is a service responsible for decoupling a caller and the handler of a Query.
type Bus interface {
	RegisterHandler(t TypeName, h Handler)
	Send(ctx context.Context, q Query) (any, error)
}

// InMemoryBus is an implementation of a query.Bus that handles queries using handlers located in memory.
type InMemoryBus struct {
	handlers map[TypeName]Handler
}

// NewInMemoryBus Creates a new instance of a InMemoryBus.
func NewInMemoryBus() *InMemoryBus {
	return &InMemoryBus{
		handlers: map[TypeName]Handler{},
	}
}

// RegisterHandler Registers a new Handler for a given Query with the bus.
func (cb *InMemoryBus) RegisterHandler(t TypeName, h Handler) {
	cb.handlers[t] = h
}

// Send a Query to its Handler for processing
func (cb *InMemoryBus) Send(ctx context.Context, q Query) (any, error) {
	// Handle
	return cb.handleQuery(q, ctx)
}

func (cb *InMemoryBus) handleQuery(q Query, ctx context.Context) (any, error) {
	handler, err := cb.resolveHandler(q)
	if err != nil {
		// Query should always be resolved! This is a critical error!
		panic(errors.Wrapf(err, "failed handling query %s", q.TypeName()))
	}
	result, err := handler.Handle(q, ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed handling query %s", q.TypeName())
	}

	return result, nil
}

func (cb *InMemoryBus) resolveHandler(q Query) (Handler, error) {
	if handler, found := cb.handlers[q.TypeName()]; !found {
		return nil, errors.Errorf("no handler found for query \"%s\"", q.TypeName())
	} else {
		return handler, nil
	}
}
