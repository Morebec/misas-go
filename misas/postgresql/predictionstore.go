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
	"database/sql"
	"encoding/json"
	"github.com/morebec/misas-go/misas"
	"github.com/morebec/misas-go/misas/clock"
	"github.com/morebec/misas-go/misas/prediction"
	"github.com/pkg/errors"
	"time"
)

type PredictionStore struct {
	connectionString string
	database         *sql.DB
	clock            clock.Clock
}

func NewPredictionStore(connectionString string, clock clock.Clock) *PredictionStore {
	return &PredictionStore{
		connectionString: connectionString,
		database:         nil,
		clock:            clock,
	}
}

func (ps *PredictionStore) setupSchemas() error {
	createTableSql := `create table if not exists predictions
(
    id                varchar(255) not null primary key,
    will_occur_at     timestamp(0) not null,
    data              jsonb        not null,
    metadata          jsonb,
    type              varchar(255) not null
);

CREATE INDEX IF NOT EXISTS idx_id
    ON predictions (id);
`

	_, err := ps.database.Exec(createTableSql)
	if err != nil {
		return errors.Wrap(err, "failed creating table predictions")
	}

	return nil
}

func (ps *PredictionStore) OpenConnection() error {
	db, err := sql.Open("postgres", ps.connectionString)
	if err != nil {
		return errors.Wrap(err, "failed opening connection to event store")
	}
	ps.database = db

	if err = ps.database.Ping(); err != nil {
		return errors.Wrap(err, "failed opening connection to event store")
	}

	return ps.setupSchemas()
}

func (ps *PredictionStore) CloseConnection() error {
	if err := ps.database.Close(); err != nil {
		return errors.Wrap(err, "failed closing connection to event store")
	}
	return nil
}

func (ps *PredictionStore) Add(p prediction.Prediction, m misas.Metadata) error {
	insertSql := "INSERT INTO predictions (id, will_occur_at, data, type, metadata) VALUES ($1, $2, $3, $4, $5)"

	predictionAsJson, err := json.Marshal(p)
	if err != nil {
		return errors.Wrap(err, "failed adding prediction to the prediction store:")
	}

	metadataAsJson, err := json.Marshal(m)
	if err != nil {
		return errors.Wrap(err, "failed adding prediction to the prediction store:")
	}

	if _, err = ps.database.Exec(insertSql, p.ID(), p.WillOccurAt(), predictionAsJson, metadataAsJson); err != nil {
		return errors.Wrapf(err, "failed adding prediction \"%s\" of type \"%s\" to prediction store", p.ID(), p.TypeName())
	}

	return nil
}

func (ps *PredictionStore) Remove(id prediction.ID) error {
	if _, err := ps.database.Exec("DELETE FROM predictions WHERE ID = $1", id); err != nil {
		return err
	}

	return nil
}

func (ps *PredictionStore) FindOccurredBefore(dt time.Time) ([]prediction.Descriptor, error) {
	rows, err := ps.database.Query("SELECT id, data, type, metadata FROM predictions WHERE will_occur_at < $1", dt)
	if err != nil {
		return nil, errors.Wrapf(err, "failed finding predictions before datetime %s", dt)
	}

	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var descriptors []prediction.Descriptor
	for rows.Next() {
		d := prediction.Descriptor{}
		var pType string
		var jsonMetadata []byte
		var jsonPredictionData []byte
		if err := rows.Scan(
			&d.ID,
			&jsonPredictionData,
			&pType,
			&jsonMetadata,
		); err != nil {
			return nil, errors.Wrap(err, "failed reading stored prediction")
		}

		var payload map[string]any
		err = json.Unmarshal(jsonPredictionData, &payload)
		if err != nil {
			return nil, errors.Wrap(err, "failed reading stored prediction")
		}

		if err := json.Unmarshal(jsonMetadata, &d.Metadata); err != nil {
			return nil, errors.Wrap(err, "failed reading stored prediction")
		}

		descriptors = append(descriptors, d)
	}

	return descriptors, nil
}

func NewPostgreSQLPredictionStore(connectionString string) *PredictionStore {
	return &PredictionStore{connectionString: connectionString}
}
