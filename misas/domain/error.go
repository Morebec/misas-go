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

import (
	"fmt"
)

// ErrorTypeName represents the kind of error in a business language specific way.
type ErrorTypeName string

// NotFoundTag is an error tag indicating that an error represents a resource that was not found when it was expected to be.
const NotFoundTag = "not_found"

// ValidationErrorTag is an error tag that indicates an error represents a validation error.
const ValidationErrorTag = "validation_error"

// AlreadyExistsTag is an error tag indicating that an error was caused by a resource existing when it wasn't expected to be.
const AlreadyExistsTag = "already_exists"

// IsDomainError Indicates if a given error is a Domain Error.
func IsDomainError(err error) bool {
	_, ok := err.(Error)
	return ok
}

// IsDomainErrorWithTypeName Indicates if a given error is a Domain Error of a certain type
func IsDomainErrorWithTypeName(err error, t ErrorTypeName) bool {
	if !IsDomainError(err) {
		return false
	}

	return err.(Error).typeName == t
}

// IsNotFoundError indicates if an error has a NotFoundTag.
func IsNotFoundError(err error) bool {
	return HasErrorTag(err, NotFoundTag)
}

// IsValidationError indicates if an error has a ValidationErrorTag.
func IsValidationError(err error) bool {
	return HasErrorTag(err, ValidationErrorTag)
}

// HasErrorTag Indicates if an error has an error tag.
func HasErrorTag(err error, tag string) bool {
	if !IsDomainError(err) {
		return false
	}

	return err.(Error).HasTag(tag)
}

// NewError Allows creating a new Error using a type and a message.
func NewError(t ErrorTypeName, opts ...ErrorOption) Error {
	err := Error{
		typeName: t,
	}

	for _, opt := range opts {
		err = opt(err)
	}

	if err.typeName == "" {
		panic("invalid call to domain.NewError: no type name provided")
	}

	return err
}

// Error represents a domain specific error, it should represent a problem that has some meaning from
// a business language point of view.
type Error struct {
	// The message of the error.
	message string
	// The kind of error
	typeName ErrorTypeName
	// the parent error that caused the current error, if any
	cause error

	// Represents additional data about this error, if any.
	data map[string]any
	tags []string
}

func (d Error) Error() string {
	return d.message + ": " + d.cause.Error()
}

func (d Error) Cause() error {
	return d.cause
}

func (d Error) TypeName() ErrorTypeName {
	return d.typeName
}

// Data returns the contextual data of this error.
func (d Error) Data() map[string]any {
	return d.data
}

// Tags returns the tags of this error.
func (d Error) Tags() []string {
	return d.tags
}

func (d Error) Unwrap() error {
	return d.cause
}

// HasTag indicates if an error has a certain tag.
func (d Error) HasTag(tag string) bool {
	for _, t := range d.tags {
		if t == tag {
			return true
		}
	}

	return false
}

type ErrorOption func(e Error) Error

// WithMessage allows specifying the message of an error.
// The message of an error should be for additional debug information that is useful for developers.
// If an error should be communicated to the user, the error's ErrorTypeName and Data should be used instead.
func WithMessage(message string) ErrorOption {
	return func(e Error) Error {
		e.message = message
		return e
	}
}

// WithMessagef allows specifying the message of an error using a formatted string.
func WithMessagef(format string, params ...any) ErrorOption {
	return WithMessage(fmt.Sprintf(format, params...))
}

// WithData allows specifying additional contextual data about an error. For example, in the case of a not found error,
// the ID of the entity that was not found could be provided as additional data.
func WithData(data map[string]any) ErrorOption {
	return func(e Error) Error {
		if e.data == nil {
			e.data = map[string]any{}
		}

		// add all the keys as data.
		for k, v := range data {
			e.data[k] = v
		}

		return e
	}
}

// WithCause allows specifying the underlying cause of a given error.
func WithCause(err error) ErrorOption {
	return func(e Error) Error {
		e.cause = err
		return e
	}
}

// WithTags WithTag allows specifying that an error has specific tags.
func WithTags(tags ...string) ErrorOption {
	return func(e Error) Error {
		e.tags = append(e.tags, tags...)
		return e
	}
}
