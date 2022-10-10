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
)

// Handler is a service responsible for executing the business logic associated with a given Command.
type Handler interface {
	// Handle a Command in a given context.Context and returns the list of changes that occurred as a result, or an error.
	Handle(ctx context.Context, c Command) (any, error)
}

// HandlerFunc Allows using a function as a Handler
type HandlerFunc func(ctx context.Context, c Command) (any, error)

func (cf HandlerFunc) Handle(ctx context.Context, c Command) (any, error) {
	return cf(ctx, c)
}
