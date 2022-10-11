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
	"github.com/morebec/go-system/misas/query"
)

type QueryHandlingOption func(s *System)

func WithQueryHandling(opts ...QueryHandlingOption) Option {
	return func(s *System) {
		for _, opt := range opts {
			opt(s)
		}
	}
}

// WithQueryBus specifies the query bus the System relies on.
func WithQueryBus(b query.Bus) QueryHandlingOption {
	return func(s *System) {
		s.QueryBus = b
	}
}

type QueryConfigurator struct {
	system *System
	query  query.Query
}

func (c QueryConfigurator) HandledBy(h query.Handler) QueryConfigurator {
	c.system.QueryBus.RegisterHandler(c.query.TypeName(), h)
	return c
}
