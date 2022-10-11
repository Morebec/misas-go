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
	"github.com/morebec/go-system/misas/command"
)

type CommandProcessingOptions func(s *System)

func WithCommandHandling(opts ...CommandProcessingOptions) Option {
	return func(s *System) {
		for _, opt := range opts {
			opt(s)
		}
	}
}

// WithCommandBus specifies the command bus the System relies on.
func WithCommandBus(b command.Bus) CommandProcessingOptions {
	return func(s *System) {
		s.CommandBus = b
	}
}

type CommandConfigurator struct {
	system  *System
	command command.Command
}

func (c CommandConfigurator) HandledBy(h command.Handler) CommandConfigurator {
	c.system.CommandBus.RegisterHandler(c.command.TypeName(), h)
	return c
}
