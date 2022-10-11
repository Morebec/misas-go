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
	"github.com/morebec/go-system/misas/prediction"
)

type PredictionHandlingOption func(system *System)

func WithPredictionHandling(opts ...PredictionHandlingOption) Option {
	return func(s *System) {
		for _, opt := range opts {
			opt(s)
		}
	}
}

func WithPredictionBus(bus prediction.Bus) PredictionHandlingOption {
	return func(s *System) {
		s.PredictionBus = bus
	}
}

// WithPredictionStore specifies the event bus that the System relies on.
func WithPredictionStore(store prediction.Store) PredictionHandlingOption {
	return func(s *System) {
		s.PredictionStore = store
	}
}

func WithUpcastingPredictionStoreDecoration() PredictionHandlingOption {
	return func(s *System) {
		if s.PredictionStore == nil {
			panic("Define the prediction store to use before indicating decoration.")
		}
		s.PredictionStore = prediction.NewUpcastingPredictionStoreDecorator(s.PredictionStore, s.PredictionUpcasterChain)
	}
}
