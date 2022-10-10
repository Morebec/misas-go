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

package postgresql

import (
	"context"
	"github.com/jwillp/go-system/misas/event/processing"
	"github.com/stretchr/testify/assert"
	"testing"
)

func buildCheckpointStore() *CheckpointStore {
	store := NewCheckpointStore("postgres://postgres@localhost:5432/postgres?sslmode=disable")

	if err := store.OpenConnection(); err != nil {
		panic(err)
	}

	if err := store.Clear(); err != nil {
		panic(err)
	}

	return store
}

func TestCheckpointStore_CloseConnection(t *testing.T) {
	store := buildCheckpointStore()
	err := store.CloseConnection()
	assert.NoError(t, err)
}

func TestCheckpointStore_FindById(t *testing.T) {
	// TODO
}

func TestCheckpointStore_OpenConnection(t *testing.T) {
	assert.NotPanics(t, func() {
		_ = buildCheckpointStore()
	})
}

func TestCheckpointStore_Remove(t *testing.T) {
	store := buildCheckpointStore()

	checkpoint := processing.Checkpoint{}
	err := store.Save(context.Background(), checkpoint)
	assert.NoError(t, err)
}

func TestCheckpointStore_Store(t *testing.T) {

}

func TestCheckpointStore_setupSchemas(t *testing.T) {

}
