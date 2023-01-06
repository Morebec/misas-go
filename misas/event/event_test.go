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

package event

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const unitTestSucceededTypeName PayloadTypeName = "unit_test.succeeded"

type unitTestSucceeded struct{}

func (u unitTestSucceeded) TypeName() PayloadTypeName {
	return unitTestSucceededTypeName
}

const unitTestFailedTypeName PayloadTypeName = "unit_test.failed"

type unitTestFailed struct{}

func (u unitTestFailed) TypeName() PayloadTypeName {
	return unitTestFailedTypeName
}

func TestEventList_Exclude(t *testing.T) {
	l := List{
		New(unitTestSucceeded{}),
		New(unitTestFailed{}),
	}

	excluded, err := l.Exclude(func(e Event) (bool, error) {
		return e.Payload.TypeName() == unitTestSucceededTypeName, nil
	})
	assert.Nil(t, err)
	assert.Len(t, excluded, 1)
	assert.Equal(t, (*excluded.First()).Payload.TypeName(), unitTestFailedTypeName)
}

func TestEventList_ExcludeByTypeNames(t *testing.T) {
	l := List{
		New(unitTestSucceeded{}),
		New(unitTestFailed{}),
	}

	excluded := l.ExcludeByTypeNames(unitTestFailedTypeName)
	assert.Len(t, excluded, 1)
	assert.Equal(t, (*excluded.First()).Payload.TypeName(), unitTestSucceededTypeName)
}

func TestEventList_IsEmpty(t *testing.T) {
	assert.False(t, List{New(unitTestSucceeded{}), New(unitTestFailed{})}.IsEmpty())
	assert.True(t, List{}.IsEmpty())
}

func TestEventList_Select(t *testing.T) {
	l := List{
		New(unitTestSucceeded{}),
		New(unitTestFailed{}),
	}

	selected, err := l.Select(func(e Event) (bool, error) {
		return e.Payload.TypeName() == unitTestSucceededTypeName, nil
	})
	assert.Nil(t, err)
	assert.Len(t, selected, 1)
	assert.Equal(t, (*selected.First()).Payload.TypeName(), unitTestSucceededTypeName)
}

func TestEventList_SelectByTypeNames(t *testing.T) {
	l := List{
		New(unitTestSucceeded{}),
		New(unitTestFailed{}),
	}

	selected := l.SelectByTypeNames(unitTestFailedTypeName)
	assert.Len(t, selected, 1)
	assert.Equal(t, (*selected.First()).Payload.TypeName(), unitTestFailedTypeName)
}

func TestEventList_First(t *testing.T) {
	assert.Equal(t, (*List{New(unitTestSucceeded{}), New(unitTestFailed{})}.First()).Payload.TypeName(), unitTestSucceededTypeName)
	assert.Nil(t, List{}.First())
}

func TestEventList_Last(t *testing.T) {
	assert.Equal(t, (*List{New(unitTestSucceeded{}), New(unitTestFailed{})}.Last()).Payload.TypeName(), unitTestFailedTypeName)
	assert.Nil(t, List{}.Last())
}
