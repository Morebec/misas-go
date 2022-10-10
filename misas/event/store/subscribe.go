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

import "github.com/pkg/errors"

// SubscribeToStreamOptions Represents the options
type SubscribeToStreamOptions struct {
	EventTypeNameFilter *TypeNameFilter
}

type SubscribeToStreamOption func(options *SubscribeToStreamOptions)

func WithSubscriptionFilter(opts ...TypeNameFilterOption) SubscribeToStreamOption {
	return func(o *SubscribeToStreamOptions) {
		if len(opts) == 0 {
			o.EventTypeNameFilter = nil
		} else {
			for _, opt := range opts {
				opt(o.EventTypeNameFilter)
			}
		}
	}
}

// Subscription allows listening to a specific Stream in order to receive notifications about new events being appended to some given streams.
type Subscription struct {
	eventChannel chan RecordedEventDescriptor
	errorChannel chan error
	close        chan<- bool
	streamID     StreamID
	options      SubscribeToStreamOptions
}

func NewSubscription(eventChannel chan RecordedEventDescriptor, errorChannel chan error, close chan<- bool, streamID StreamID, options SubscribeToStreamOptions) *Subscription {
	return &Subscription{eventChannel: eventChannel, errorChannel: errorChannel, close: close, streamID: streamID, options: options}
}

func (s Subscription) Options() SubscribeToStreamOptions {
	return s.options
}

func (s Subscription) EventChannel() <-chan RecordedEventDescriptor {
	return s.eventChannel
}

func (s Subscription) ErrorChannel() <-chan error {
	return s.errorChannel
}

// Listen Helper function allowing to listen to events and errors using callbacks. This method is blocking.
func (s Subscription) Listen(eventFunc func(d RecordedEventDescriptor) error, errorFunc func(err error) error) error {
	for {
		select {
		case d := <-s.eventChannel:
			if err := eventFunc(d); err != nil {
				return errors.Wrap(err, "failed listening to subscription events")
			}
		case err := <-s.errorChannel:
			if err = errorFunc(err); err != nil {
				return errors.Wrap(err, "failed listening to subscription events")
			}
		}
	}
}

// EmitEvent emits an RecordedEventDescriptor to this subscription. This method is intended to be used by EventStore implementations.
func (s Subscription) EmitEvent(d RecordedEventDescriptor) {
	s.eventChannel <- d
}

// EmitError emits an error to this subscription. This method is intended to be used by EventStore implementations.
func (s Subscription) EmitError(err error) {
	s.errorChannel <- err
}

func (s Subscription) StreamID() StreamID {
	return s.streamID
}

func (s Subscription) Close() error {
	s.close <- true
	return nil
}

func BuildSubscribeToStreamOptions(opts []SubscribeToStreamOption) SubscribeToStreamOptions {
	options := &SubscribeToStreamOptions{
		EventTypeNameFilter: nil,
	}
	for _, opt := range opts {
		opt(options)
	}
	return *options
}
