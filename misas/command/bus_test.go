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
	"github.com/stretchr/testify/assert"
	"testing"
)

const runUnitTestCommandTypeName PayloadTypeName = "unit_test.run"

type runUnitTestCommandPayload struct {
}

func (r runUnitTestCommandPayload) TypeName() PayloadTypeName {
	return runUnitTestCommandTypeName
}

type runUnitTestCommandHandler struct {
}

func (r runUnitTestCommandHandler) Handle(context.Context, Command) (any, error) {
	return nil, nil
}

func TestInMemoryBus_RegisterHandler(t *testing.T) {
	bus := NewInMemoryBus()
	bus.RegisterHandler(runUnitTestCommandTypeName, runUnitTestCommandHandler{})
}

func TestInMemoryBus_Send(t *testing.T) {
	bus := NewInMemoryBus()
	bus.RegisterHandler(runUnitTestCommandTypeName, runUnitTestCommandHandler{})

	events, err := bus.Send(context.Background(), New(runUnitTestCommandPayload{}))
	assert.Nil(t, err)
	assert.Nil(t, events)
}

func TestNewInMemoryBus(t *testing.T) {
	assert.NotNil(t, NewInMemoryBus())
}
