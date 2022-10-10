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
	"database/sql"
	"github.com/jwillp/go-system/misas/event/processing"
	"github.com/pkg/errors"
)

// CheckpointStore is a PostgreSQL implementation of a checkpoint store in a table named "checkpoints"
type CheckpointStore struct {
	connectionString string
	database         *sql.DB
}

func NewCheckpointStore(connectionString string) *CheckpointStore {
	return &CheckpointStore{connectionString: connectionString}
}

func (p *CheckpointStore) setupSchemas() error {
	createTableCheckpointSql := `
CREATE TABLE IF NOT EXISTS checkpoints
(
    id        VARCHAR(255) NOT NULL PRIMARY KEY,
    stream_id VARCHAR(255) NOT NULL,
    position  INTEGER NOT NULL
);`

	_, err := p.database.Exec(createTableCheckpointSql)
	if err != nil {
		return errors.Wrap(err, "failed creating table checkpoints")
	}

	return nil
}

func (p *CheckpointStore) OpenConnection() error {
	db, err := sql.Open("postgres", p.connectionString)
	if err != nil {
		return errors.Wrap(err, "failed opening connection to event store")
	}
	p.database = db

	if err = p.database.Ping(); err != nil {
		return errors.Wrap(err, "failed opening connection to event store")
	}

	return p.setupSchemas()
}

func (p *CheckpointStore) CloseConnection() error {
	if err := p.database.Close(); err != nil {
		return errors.Wrap(err, "failed closing connection to event store")
	}
	return nil
}

func (p *CheckpointStore) Save(ctx context.Context, checkpoint processing.Checkpoint) error {

	insertSql := `
INSERT INTO checkpoints (id, stream_id, position) 
VALUES($1, $2, $3);
`

	_, err := p.database.Exec(insertSql, checkpoint.ID, checkpoint.StreamID, checkpoint.Position)
	if err != nil {
		return errors.Wrapf(err,
			"failed storing checkpoint \"%s\" for stream \"%s\"",
			checkpoint.ID,
			checkpoint.StreamID,
		)
	}

	return nil
}

func (p *CheckpointStore) FindById(ctx context.Context, id processing.CheckpointID) (*processing.Checkpoint, error) {
	selecSql := `
SELECT id, stream_id, position FROM checkpoints
WHERE id = $1;
`
	row := p.database.QueryRow(selecSql, id)
	if row.Err() != nil {
		return nil, errors.Wrapf(row.Err(), "failed retrieving checkpoint \"%s\"", id)
	}

	checkpoint := &processing.Checkpoint{}
	if err := row.Scan(
		&checkpoint.ID,
		&checkpoint.StreamID,
		&checkpoint.Position,
	); err != nil {
		return nil, errors.Wrapf(row.Err(), "failed retrieving checkpoint \"%s\"", id)
	}

	return checkpoint, nil

}

func (p *CheckpointStore) Remove(ctx context.Context, id processing.CheckpointID) error {
	deleteSql := `
DELETE FROM checkpoints
WHERE id = $1
;
`
	if _, err := p.database.Exec(deleteSql, id); err != nil {
		return errors.Wrapf(err, "failed removing checkpoint \"%s\"", id)
	}

	return nil
}

func (p *CheckpointStore) Clear() error {
	if _, err := p.database.Exec("TRUNCATE TABLE checkpoints"); err != nil {
		return errors.Wrap(err, "failed clearing checkpoint store")
	}

	return nil
}
