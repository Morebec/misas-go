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

import (
	"github.com/morebec/misas-go/misas/event"
	"time"
)

// TruncateStreamOptions Options for truncating a stream. All events that have a version number smaller than the BeforePosition will be truncated.
type TruncateStreamOptions struct {
	BeforePosition Position
	Reason         *string
}

type TruncateStreamOption func(options *TruncateStreamOptions)

func BeforePosition(p Position) TruncateStreamOption {
	return func(options *TruncateStreamOptions) {
		options.BeforePosition = p
	}
}

const StreamTruncatedEventTypeName event.PayloadTypeName = "es.stream.truncated"

type StreamTruncatedEvent struct {
	StreamID    string
	Reason      *string
	TruncatedAt time.Time
}

func (s StreamTruncatedEvent) TypeName() event.PayloadTypeName {
	return StreamTruncatedEventTypeName
}

func BuildTruncateFromStreamOptions(opts []TruncateStreamOption) TruncateStreamOptions {
	options := &TruncateStreamOptions{}

	for _, opt := range opts {
		opt(options)
	}

	return *options
}
