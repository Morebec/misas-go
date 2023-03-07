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
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/morebec/misas-go/misas"
	"github.com/morebec/misas-go/misas/clock"
	"github.com/morebec/misas-go/misas/event/store"
	"github.com/pkg/errors"
	"strings"
	"time"
)

const GlobalStreamID store.StreamID = "$all"

const InternalStreamID = "$es"

type EventStore struct {
	connectionString string
	database         *sql.DB
	clock            clock.Clock
	upcasterChain    store.UpcasterChain
}

func NewEventStore(
	connectionString string,
	clock clock.Clock,
) *EventStore {
	return &EventStore{
		connectionString: connectionString,
		database:         nil,
		clock:            clock,
	}
}

func (es *EventStore) setupSchemas(ctx context.Context) error {
	createTableEventsSql := `
CREATE TABLE IF NOT EXISTS events 
(
    id              VARCHAR(255) NOT NULL,
    stream_id       VARCHAR(255) NOT NULL,
    stream_version  INTEGER      NOT NULL,
    type            VARCHAR(255) NOT NULL,
    metadata        JSONB        NOT NULL,
    data            JSONB        NOT NULL,
    recorded_at     TIMESTAMP(0) NOT NULL,
    sequence_number SERIAL
);

CREATE INDEX IF NOT EXISTS idx_id
    ON events (id);

CREATE INDEX IF NOT EXISTS idx_stream_id
    ON events (stream_id);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_id_stream_id
    ON events (id, stream_id);

CREATE INDEX IF NOT EXISTS idx_stream_version
    ON events (stream_version);

CREATE INDEX IF NOT EXISTS idx_sequence_number
    ON events (sequence_number);
`
	_, err := es.database.ExecContext(ctx, createTableEventsSql)
	if err != nil {
		return errors.Wrap(err, "failed creating table events")
	}

	createStreamsTableSql := `
CREATE TABLE IF NOT EXISTS streams
(
    id      VARCHAR(255)      NOT NULL PRIMARY KEY,
    version INTEGER DEFAULT 0 NOT NULL
);
`
	_, err = es.database.ExecContext(ctx, createStreamsTableSql)
	if err != nil {
		return errors.Wrap(err, "failed creating table streams")
	}

	notifyEventsSql := `
-- Create the trigger function
CREATE OR REPLACE FUNCTION notify_events() RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('events', row_to_json(NEW)::text);
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

-- Create the trigger
DROP TRIGGER IF EXISTS notify_events_trigger ON events;
CREATE TRIGGER notify_events_trigger
AFTER INSERT ON events
FOR EACH ROW EXECUTE PROCEDURE notify_events();
`

	_, err = es.database.ExecContext(ctx, notifyEventsSql)
	if err != nil {
		return errors.Wrap(err, "failed creating notification and function trigger")
	}

	return nil
}

func (es *EventStore) Open(ctx context.Context) error {
	db, err := sql.Open("postgres", es.connectionString)
	if err != nil {
		return errors.Wrap(err, "failed opening connection to event store")
	}
	es.database = db

	if err = es.database.PingContext(ctx); err != nil {
		return errors.Wrap(err, "failed opening connection to event store")
	}

	return es.setupSchemas(ctx)
}

func (es *EventStore) Close() error {
	if err := es.database.Close(); err != nil {
		return errors.Wrap(err, "failed closing connection to event store")
	}
	return nil
}

func (es *EventStore) GlobalStreamID() store.StreamID {
	return GlobalStreamID
}

func (es *EventStore) AppendToStream(ctx context.Context, streamID store.StreamID, events []store.EventDescriptor, opts ...store.AppendToStreamOption) error {

	options := store.BuildAppendToStreamOptions(opts)

	// Ensure it is not a virtual stream
	if streamID == es.GlobalStreamID() {
		return errors.Errorf("cannot append to virtual stream \"%s\"", streamID)
	}

	if len(events) == 0 {
		return nil
	}

	stream, err := es.GetStream(ctx, streamID)
	streamFound := true
	if err != nil {
		if errors.Is(err, store.NewStreamNotFoundError(streamID)) {
			streamFound = false
		} else {
			return errors.Wrapf(err, "failed appending to stream \"%s\"", streamID)
		}
	}

	var streamVersion store.StreamVersion
	if streamFound {
		streamVersion = stream.Version
	} else {
		streamVersion = store.InitialVersion
	}

	// Check concurrency
	if options.ExpectedVersion != nil && *options.ExpectedVersion != streamVersion {
		return store.NewConcurrencyError(streamID, *options.ExpectedVersion, streamVersion)
	}

	tx, err := es.database.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrapf(err, "failed starting transaction when appending events to stream \"%s\"", streamID)
	}

	for _, d := range events {
		streamVersion++

		insertEventSql := `
INSERT INTO events (id, stream_id, stream_version, type, metadata, data, recorded_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`
		eventAsJson, err := json.Marshal(d.Payload)
		if err != nil {
			return errors.Wrap(err, "failed appending events to the event store")
		}

		metadataAsJson, err := json.Marshal(d.Metadata)
		if err != nil {
			return errors.Wrap(err, "failed appending events to the event store")
		}

		if _, err = tx.ExecContext(ctx, insertEventSql, d.ID, streamID, streamVersion, d.TypeName, metadataAsJson, eventAsJson, es.clock.Now()); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return errors.Wrap(rollbackErr, "failed rolling back transaction when appending event to the event store")
			}
			return errors.Wrap(err, "failed appending event to the event store")
		}
	}

	if err = es.updateStreamVersionIndex(ctx, tx, streamID, streamVersion); err != nil {
		return errors.Wrap(err, "failed appending event to the event store")
	}

	if err = tx.Commit(); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrap(rollbackErr, "failed rolling back transaction when appending events to the event store")
		}
		return errors.Wrap(err, "failed appending events to the event store")
	}

	return nil
}

func (es *EventStore) ReadFromStream(ctx context.Context, streamID store.StreamID, opts ...store.ReadFromStreamOption) (store.StreamSlice, error) {
	options := store.BuildReadFromStreamOptions(opts)
	isGlobalStream := streamID == es.GlobalStreamID()
	if !isGlobalStream {
		streamExists, err := es.StreamExists(ctx, streamID)
		if err != nil {
			return store.StreamSlice{}, errors.Wrapf(err, "failed reading from stream \"%s\"", streamID)
		}

		if !streamExists {
			return store.StreamSlice{}, store.NewStreamNotFoundError(streamID)
		}
	}

	var stmtParams []any
	stmtParamCounter := 1
	var whereClauses []string

	selectSql := "SELECT id, type, stream_id, stream_version, data, metadata, sequence_number, recorded_at FROM events"

	if !isGlobalStream {
		whereClauses = append(whereClauses, fmt.Sprintf("stream_id = $%d", stmtParamCounter))
		stmtParamCounter++
		stmtParams = append(stmtParams, streamID)
	}

	if options.Position >= store.Position(store.InitialVersion) {
		var positionColumn string
		if isGlobalStream {
			positionColumn = "sequence_number"
		} else {
			positionColumn = "stream_version"
		}
		var positionValue string
		if options.Position == store.End {
			if isGlobalStream {
				positionValue = fmt.Sprintf("(SELECT MAX(%s) + 1 FROM events)", positionColumn)
			} else {
				positionValue = fmt.Sprintf("(SELECT MAX(%s) + 1 FROM events WHERE stream_id = $%d)", positionColumn, stmtParamCounter)
				stmtParams = append(stmtParams, fmt.Sprintf("%s", streamID))
				stmtParamCounter++
			}
		} else {
			positionValue = fmt.Sprintf("$%d", stmtParamCounter)
			stmtParams = append(stmtParams, fmt.Sprintf("%d", options.Position))
			stmtParamCounter++
		}

		var positionSign string
		if options.Direction == store.Forward {
			positionSign = ">"
		} else {
			positionSign = "<"
		}

		whereClauses = append(whereClauses, fmt.Sprintf("%s %s %s", positionColumn, positionSign, positionValue))
	}

	var direction string
	if options.Direction == store.Forward {
		direction = "ASC"
	} else {
		direction = "DESC"
	}
	orderBySql := fmt.Sprintf("ORDER BY sequence_number %s", direction)

	var limitSql string
	if options.MaxCount > 0 {
		limitSql = fmt.Sprintf("LIMIT %d", options.MaxCount)
	} else {
		limitSql = ""
	}

	querySql := fmt.Sprintf(`
%s
WHERE %s
%s
%s
`, selectSql, strings.Join(whereClauses, " AND "), orderBySql, limitSql)

	rows, err := es.database.QueryContext(ctx, querySql, stmtParams...)
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	if err != nil {
		return store.StreamSlice{}, errors.Wrapf(err, "failed reading from stream \"%s\"", streamID)
	}

	streamSlice := store.StreamSlice{
		StreamID:    streamID,
		Descriptors: []store.RecordedEventDescriptor{},
	}
	for rows.Next() {
		var descriptor store.RecordedEventDescriptor
		var jsonEventData []byte
		var jsonMetadata []byte

		if err := rows.Scan(
			&descriptor.ID,
			&descriptor.TypeName,
			&descriptor.StreamID,
			&descriptor.Version,
			&jsonEventData,
			&jsonMetadata,
			&descriptor.SequenceNumber,
			&descriptor.RecordedAt,
		); err != nil {
			return store.StreamSlice{}, errors.Wrapf(err, "failed reading from stream \"%s\"", streamID)
		}

		if options.EventTypeNameFilter != nil {
			if options.EventTypeNameFilter.Mode == store.Exclude {
				for _, tn := range options.EventTypeNameFilter.EventTypeNames {
					if descriptor.TypeName == tn {
						continue
					}
				}
			}

			if options.EventTypeNameFilter.Mode == store.Select {
				for _, tn := range options.EventTypeNameFilter.EventTypeNames {
					if descriptor.TypeName != tn {
						continue
					}
				}
			}
		}

		if err := json.Unmarshal(jsonEventData, &descriptor.Payload); err != nil {
			return store.StreamSlice{}, errors.Wrapf(err, "failed reading from stream \"%s\"", streamID)
		}

		if err := json.Unmarshal(jsonMetadata, &descriptor.Metadata); err != nil {
			return store.StreamSlice{}, errors.Wrapf(err, "failed reading from stream \"%s\"", streamID)
		}

		streamSlice.Descriptors = append(streamSlice.Descriptors, descriptor)
	}

	if err = rows.Err(); err != nil {
		return store.StreamSlice{}, errors.Wrapf(err, "failed reading from stream \"%s\"", streamID)
	}

	defer func(result *sql.Rows) {
		err := result.Close()
		if err != nil {
			// TODO
			panic(err)
		}
	}(rows)

	return streamSlice, nil
}

func (es *EventStore) TruncateStream(ctx context.Context, id store.StreamID, opts ...store.TruncateStreamOption) error {
	options := store.BuildTruncateFromStreamOptions(opts)

	tx, err := es.database.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrapf(err, "failed truncating from stream \"%s\"", id)
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM events WHERE stream_id = $1 AND stream_version < $2", id, options.BeforePosition)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.Wrapf(err, "failed rolling back transaction when truncating stream \"%s\"", id)
		}
		return errors.Wrapf(err, "failed truncating from stream \"%s\"", id)
	}

	err = es.AppendToStream(ctx, InternalStreamID, []store.EventDescriptor{
		{
			ID:       store.EventID(uuid.New().String()),
			TypeName: store.StreamTruncatedEventTypeName,
			Payload: store.DescriptorPayload{
				"streamId":    "",
				"reason":      options.Reason,
				"truncatedAt": es.clock.Now(),
			},
			Metadata: misas.Metadata{},
		},
	}, store.WithOptimisticConcurrencyCheckDisabled())
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.Wrapf(err, "failed rolling back transaction when truncating stream \"%s\"", id)
		}
		return errors.Wrapf(err, "failed truncating from stream \"%s\"", id)
	}

	if err = tx.Commit(); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrapf(rollbackErr, "failed rolling back transaction when truncating stream \"%s\"", id)
		}
		return errors.Wrapf(err, "failed truncating from stream \"%s\"", id)
	}

	return nil
}

func (es *EventStore) DeleteStream(ctx context.Context, id store.StreamID) error {
	tx, err := es.database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, "DELETE FROM events WHERE stream_id = $1", id); err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.Wrapf(err, "failed rolling back transaction when deleting stream \"%s\"", id)
		}
		return errors.Wrapf(err, "failed deleting stream \"%s\"", id)
	}

	if _, err = tx.ExecContext(ctx, "DELETE FROM streams WHERE id = $1", id); err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.Wrapf(err, "failed rolling back transaction when deleting stream \"%s\"", id)
		}
		return errors.Wrapf(err, "failed deleting stream \"%s\"", id)
	}

	err = es.AppendToStream(ctx, InternalStreamID, []store.EventDescriptor{
		{
			ID:       store.EventID(uuid.New().String()),
			TypeName: store.StreamTruncatedEventTypeName,
			Payload: store.DescriptorPayload{
				"streamId":  string(id),
				"reason":    nil,
				"deletedAt": es.clock.Now(),
			},
			Metadata: misas.Metadata{},
		},
	}, store.WithOptimisticConcurrencyCheckDisabled())

	if err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.Wrapf(err, "failed rolling back transaction when deleting stream \"%s\"", id)
		}
		return errors.Wrapf(err, "failed deleting stream \"%s\"", id)
	}

	err = tx.Commit()
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}

	return nil
}

func (es *EventStore) SubscribeToStream(ctx context.Context, streamID store.StreamID, opts ...store.SubscribeToStreamOption) (store.Subscription, error) {
	options := store.BuildSubscribeToStreamOptions(opts)
	errorChannel := make(chan error)
	eventChannel := make(chan store.RecordedEventDescriptor)
	closeChannel := make(chan bool, 1)
	subscription := store.NewSubscription(eventChannel, errorChannel, closeChannel, streamID, options)

	listener := pq.NewListener(es.connectionString, 5*time.Second, time.Minute, func(event pq.ListenerEventType, err error) {
		if err != nil {
			errorChannel <- err
		}
	})

	go func() {
		for {
			if err := listener.Listen("events"); err != nil {
				if err != pq.ErrChannelAlreadyOpen {
					subscription.EmitError(err)
				}
				return
			}

			select {
			case n := <-listener.Notify:
				var descriptorData map[string]any
				err := json.Unmarshal([]byte(n.Extra), &descriptorData)
				if err != nil {
					subscription.EmitError(err)
					return
				}

				// Read the event from the store and publish it on the channel.
				stream, err := es.ReadFromStream(
					ctx,
					es.GlobalStreamID(),
					store.From(store.Position(descriptorData["sequence_number"].(float64))-1),
					store.InForwardDirection(),
					store.WithMaxCount(1),
				)
				if err != nil {
					return
				}

				subscription.EmitEvent(stream.First())

			case _ = <-closeChannel:
				err := listener.Unlisten("events")
				if err != nil {
					subscription.EmitError(err)
					return
				}
			}
		}
	}()

	return *subscription, nil
}

func (es *EventStore) StreamExists(ctx context.Context, id store.StreamID) (bool, error) {
	if _, err := es.GetStream(ctx, id); err != nil {
		if errors.Is(err, store.NewStreamNotFoundError(id)) {
			return false, nil
		}
		return false, errors.Wrapf(err, "failed checking existance of stream \"%s\"", id)
	}
	return true, nil
}

func (es *EventStore) GetStream(ctx context.Context, id store.StreamID) (store.Stream, error) {
	row := es.database.QueryRowContext(ctx, "SELECT version FROM streams WHERE id = $1", id)
	if err := row.Err(); err != nil {
		return store.Stream{}, errors.Wrapf(err, "failed checking information of stream \"%s\"", id)
	}

	var streamVersion store.StreamVersion
	if err := row.Scan(&streamVersion); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.Stream{}, store.NewStreamNotFoundError(id)
		}
		return store.Stream{}, errors.Wrapf(err, "failed checking information of stream \"%s\"", id)
	}

	return store.Stream{
		ID:             id,
		Version:        streamVersion,
		InitialVersion: 0,
	}, nil
}

func (es *EventStore) Clear(ctx context.Context) error {
	tx, err := es.database.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed clearing event store")
	}

	if _, err := tx.ExecContext(ctx, "TRUNCATE TABLE events"); err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.Wrap(err, "failed rolling back transaction when clearing event store")
		}
		return errors.Wrap(err, "failed clearing event store")
	}

	if _, err := tx.ExecContext(ctx, "TRUNCATE TABLE streams"); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrap(err, "failed rolling back transaction when clearing event store")
		}
		return errors.Wrap(err, "failed clearing event store")
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (es *EventStore) updateStreamVersionIndex(ctx context.Context, tx *sql.Tx, id store.StreamID, version store.StreamVersion) error {
	upsertSql := `
INSERT INTO streams (id, version) 
VALUES($1, $2)
ON CONFLICT (id)
DO UPDATE SET id = $1, version = $2;
`
	if _, err := tx.ExecContext(ctx, upsertSql, id, version); err != nil {
		return errors.Wrap(err, "failed updating stream version index")
	}

	return nil
}
