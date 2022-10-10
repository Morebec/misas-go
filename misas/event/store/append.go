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

package store

import "fmt"

type ConcurrencyError struct {
	StreamID        StreamID
	ExpectedVersion StreamVersion
	ActualVersion   StreamVersion
}

func (s ConcurrencyError) Error() string {
	return fmt.Sprintf(
		`concurrency issue encountered on stream with "%s", expected version: "%d", actual version: "%d"`,
		s.StreamID,
		s.ExpectedVersion,
		s.ActualVersion,
	)
}

func NewConcurrencyError(streamID StreamID, expectedVersion StreamVersion, actualVersion StreamVersion) error {
	return ConcurrencyError{
		StreamID:        streamID,
		ExpectedVersion: expectedVersion,
		ActualVersion:   actualVersion,
	}
}

// AppendToStreamOptions represents options to alter the behaviour of the AppendsToStream function of the event store.
type AppendToStreamOptions struct {
	ExpectedVersion *StreamVersion
}

func BuildAppendToStreamOptions(opts []AppendToStreamOption) AppendToStreamOptions {
	options := &AppendToStreamOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return *options
}

type AppendToStreamOption func(options *AppendToStreamOptions)

// WithExpectedVersion Allows specifying the expected version of the stream before appending. This allows Optimistic Concurrency Checks.
func WithExpectedVersion(v StreamVersion) AppendToStreamOption {
	return func(options *AppendToStreamOptions) {
		options.ExpectedVersion = &v
	}
}

// WithOptimisticConcurrencyCheckDisabled Allows specifying that optimistic concurrency checks should not be performed, while appending.
func WithOptimisticConcurrencyCheckDisabled() AppendToStreamOption {
	return func(options *AppendToStreamOptions) {
		options.ExpectedVersion = nil
	}
}
