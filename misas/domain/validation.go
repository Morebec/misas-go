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

package domain

import "github.com/morebec/go-system/misas/command"

const InvalidCommand ErrorTypeName = "invalid_command"
const InvalidQuery ErrorTypeName = "invalid_query"

// NewInvalidCommandError helper function allowing to create a new Error with InvalidCommand as its ErrorTypeName and provided options
func NewInvalidCommandError(c command.Command, opts ...ErrorOption) Error {
	return NewError(append(opts,
		WithTypeName(InvalidCommand),
		WithMessage("invalid command"),
	)...)
}

// NewInvalidQueryError helper function allowing to create a new Error with InvalidQuery as its ErrorTypeName and provided options
func NewInvalidQueryError(opts ...ErrorOption) Error {
	return NewError(append(opts,
		WithTypeName(InvalidQuery),
		WithMessage("invalid query"),
	)...)
}

// WithInvalidField Allows specifying that a given field was invalid for a command or a query
func WithInvalidField(fieldName string, message string) ErrorOption {
	return WithData(map[string]any{
		fieldName: message,
	})
}
