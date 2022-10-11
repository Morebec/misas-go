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
	"github.com/morebec/misas-go/misas/event"
	"github.com/morebec/misas-go/misas/event/store"
)

type EventHandlingOption func(s *System)

func WithEventHandling(opts ...EventHandlingOption) Option {
	return func(s *System) {
		for _, opt := range opts {
			opt(s)
		}
	}
}

// WithEventBus specifies the event bus the System relies on.
func WithEventBus(b event.Bus) EventHandlingOption {
	return func(s *System) {
		s.EventBus = b
	}
}

// WithEventStore specifies that the System should use a given event.Save
func WithEventStore(es store.EventStore) EventHandlingOption {
	return func(s *System) {
		s.EventStore = es
	}
}

func WithUpcastingEventStoreDecoration() EventHandlingOption {
	return func(s *System) {
		if s.EventStore == nil {
			panic("Define the event store to use before indicating decoration.")
		}
		s.EventStore = store.NewUpcastingEventStoreDecorator(s.EventStore, s.EventUpcasterChain)
	}
}

func WithEventConverter(c *store.EventConverter) EventHandlingOption {
	return func(s *System) {
		s.EventConverter = c
	}
}
