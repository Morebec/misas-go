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

package event

import (
	"context"
	"github.com/pkg/errors"
)

type InMemoryBus struct {
	handlers map[TypeName][]Handler
}

func NewInMemoryBus() *InMemoryBus {
	return &InMemoryBus{handlers: map[TypeName][]Handler{}}
}

func (eb *InMemoryBus) Send(ctx context.Context, e Event) error {
	handlers := eb.resolveHandlers(e)

	for _, h := range handlers {
		if err := h.Handle(e, ctx); err != nil {
			return errors.Wrapf(err, "failed handling event \"%s\"", e.TypeName())
		}
	}

	return nil
}

func (eb *InMemoryBus) RegisterHandler(t TypeName, h Handler) {
	if _, found := eb.handlers[t]; !found {
		eb.handlers[t] = []Handler{}
	}
	eb.handlers[t] = append(eb.handlers[t], h)
}

func (eb *InMemoryBus) resolveHandlers(e Event) []Handler {
	if handlers, found := eb.handlers[e.TypeName()]; !found {
		return nil
	} else {
		return handlers
	}
}
