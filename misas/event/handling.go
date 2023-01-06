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
)

type Handler interface {
	// Handle an event in a given context
	Handle(ctx context.Context, e Event) error
}

// HandlerFunc Allows using a function as a Handler
type HandlerFunc func(ctx context.Context, e Event) error

func (ef HandlerFunc) Handle(ctx context.Context, e Event) error {
	return ef(ctx, e)
}

type Bus interface {
	// Send an event to its registered handlers
	Send(ctx context.Context, e Event) error

	// RegisterHandler Registers an event handler with this bus.
	RegisterHandler(t PayloadTypeName, h Handler)
}
