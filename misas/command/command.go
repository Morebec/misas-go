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
	"github.com/morebec/misas-go/misas"
)

// PayloadTypeName Represents the unique type name of a Command's Payload for serialization and discriminatory purposes.
type PayloadTypeName string

// Command intent by an agent (human or machine) to perform a state change in a system.
type Command struct {
	Payload  Payload
	Metadata misas.Metadata
}

type Payload interface {
	TypeName() PayloadTypeName
}

// New returns a command with a given payload.
func New(p Payload) Command {
	return NewWithMetadata(p, nil)
}

// NewWithMetadata returns a command with a payload and some given metadata
func NewWithMetadata(p Payload, m misas.Metadata) Command {
	return Command{
		Payload:  p,
		Metadata: m,
	}
}

// Handler is a service responsible for executing the business logic associated with a given Command.
type Handler interface {
	// Handle a Command in a given context.Context and returns an optional response pertaining to the handling
	// of the command, or an error if a problem occurred.
	// Command handlers are responsible for defining the type of response they return.
	Handle(ctx context.Context, c Command) (any, error)
}

// HandlerFunc Allows using a function as a Handler
type HandlerFunc func(ctx context.Context, c Command) (any, error)

func (cf HandlerFunc) Handle(ctx context.Context, c Command) (any, error) {
	return cf(ctx, c)
}
