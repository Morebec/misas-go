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

func TestHandlerFunc_Handle(t *testing.T) {
	v := 0

	h := HandlerFunc(func(ctx context.Context, c Command) (any, error) {
		v = 5

		return nil, nil
	})

	_, _ = h.Handle(context.Background(), nil)

	assert.Equal(t, 5, v)
}
