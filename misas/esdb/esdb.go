package esdb

import (
	"context"
	"encoding/json"
	"github.com/EventStore/EventStore-Client-Go/v3/esdb"
	uuid "github.com/gofrs/uuid"
	"github.com/morebec/misas-go/misas/event"
	"github.com/morebec/misas-go/misas/event/store"
	"github.com/pkg/errors"
	"io"
	"math"
	"strings"
)

const GlobalStreamID = "$all"

// EventStore is an implementation of an event store based on an eventstore.com client.
type EventStore struct {
	configuration *esdb.Configuration
	client        *esdb.Client
}

func NewEventStore(configuration *esdb.Configuration) *EventStore {
	return &EventStore{
		configuration: configuration,
	}
}

func NewEventStoreFromConnectionString(connectionString string) (*EventStore, error) {
	config, err := esdb.ParseConnectionString(connectionString)
	if err != nil {
		return nil, err
	}
	return NewEventStore(config), nil
}

func (es *EventStore) Open(ctx context.Context) error {
	client, err := esdb.NewClient(es.configuration)
	if err != nil {
		return errors.Wrapf(err, "failed opening connection to event store")
	}
	es.client = client

	return nil
}

func (es *EventStore) Close(context.Context) error {
	if err := es.client.Close(); err != nil {
		return errors.Wrapf(err, "failed closing connection to event store")
	}

	return nil
}

// Client returns the esdb.Client.
func (es *EventStore) Client() *esdb.Client {
	return es.client
}

func (es *EventStore) GlobalStreamID() store.StreamID {
	return GlobalStreamID
}

// AppendToStream appends events to a stream. EventStoreDB imposes a limitation that Event Identifiers must be UUIDv4.
func (es *EventStore) AppendToStream(ctx context.Context, streamID store.StreamID, events []store.EventDescriptor, opts ...store.AppendToStreamOption) error {
	options := store.BuildAppendToStreamOptions(opts)

	var revision esdb.ExpectedRevision
	if options.ExpectedVersion == nil {
		revision = esdb.Any{}
	} else if *options.ExpectedVersion == store.InitialVersion {
		revision = esdb.NoStream{}
	} else {
		revision = esdb.StreamExists{}
	}

	var eventData []esdb.EventData
	for _, e := range events {
		jsonPayload, err := json.Marshal(e.Payload)
		if err != nil {
			return err
		}

		jsonMetadata, err := json.Marshal(e.Metadata)
		if err != nil {
			return err
		}

		uid, err := uuid.FromString(string(e.ID))
		if err != nil {
			return err
		}
		eventData = append(eventData, esdb.EventData{
			EventID:     uid,
			EventType:   string(e.TypeName),
			ContentType: esdb.ContentTypeJson,
			Data:        jsonPayload,
			Metadata:    jsonMetadata,
		})
	}

	_, err := es.client.AppendToStream(ctx, string(streamID), esdb.AppendToStreamOptions{
		ExpectedRevision: revision,
		Authenticated:    nil,
		Deadline:         nil,
		RequiresLeader:   false,
	}, eventData...)

	if err != nil {
		return err
	}

	return nil
}

func (es *EventStore) ReadFromStream(ctx context.Context, streamID store.StreamID, opts ...store.ReadFromStreamOption) (store.StreamSlice, error) {
	options := store.BuildReadFromStreamOptions(opts)

	var dir esdb.Direction
	switch options.Direction {
	case store.Forward:
		dir = esdb.Forwards
	case store.Backward:
		dir = esdb.Backwards
	}

	var count uint64
	switch options.MaxCount {
	// there is currently a bug in esdb go client, where a value of 0 causes a read error, when it should be accepted
	// as no max
	case 0:
		count = math.MaxUint64
	default:
		count = uint64(options.MaxCount)
	}

	var stream *esdb.ReadStream
	var err error
	if streamID == es.GlobalStreamID() {
		stream, err = es.readAll(ctx, options, dir, count)
	} else {
		stream, err = es.readStream(ctx, streamID, options, dir, count)
	}
	if err != nil {
		return store.StreamSlice{}, err
	}

	defer stream.Close()

	slice := store.StreamSlice{
		StreamID: streamID,
	}
	for {
		evt, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}

		// Skip esdb system events.
		if strings.HasPrefix(evt.OriginalEvent().EventType, "$") {
			continue
		}

		if err != nil {
			return store.StreamSlice{}, err
		}

		var data map[string]any
		if err := json.Unmarshal(evt.Event.Data, &data); err != nil {
			return store.StreamSlice{}, err
		}

		var metadata map[string]any
		if err := json.Unmarshal(evt.Event.UserMetadata, &metadata); err != nil {
			return store.StreamSlice{}, err
		}

		slice.Descriptors = append(slice.Descriptors, store.RecordedEventDescriptor{
			ID:             store.EventID(evt.Event.EventID.String()),
			TypeName:       event.TypeName(evt.Event.EventType),
			Payload:        data,
			Metadata:       metadata,
			StreamID:       store.StreamID(evt.Event.StreamID),
			Version:        store.StreamVersion(evt.Event.Position.Commit),
			SequenceNumber: store.SequenceNumber(evt.Event.EventNumber),
			RecordedAt:     evt.Event.CreatedDate,
		})
	}

	return slice, nil
}

func (es *EventStore) readStream(
	ctx context.Context,
	streamID store.StreamID,
	options *store.ReadFromStreamOptions,
	dir esdb.Direction,
	count uint64,
) (*esdb.ReadStream, error) {
	var fromPosition esdb.StreamPosition
	switch options.Position {
	case store.Start:
		fromPosition = esdb.Start{}
	case store.End:
		fromPosition = esdb.End{}
	default:
		fromPosition = esdb.StreamRevision{Value: uint64(options.Position)}
	}
	return es.client.ReadStream(ctx, string(streamID), esdb.ReadStreamOptions{
		Direction:      dir,
		From:           fromPosition,
		ResolveLinkTos: false,
		Authenticated:  nil,
		Deadline:       nil,
		RequiresLeader: false,
	}, count)
}

func (es *EventStore) readAll(
	ctx context.Context,
	options *store.ReadFromStreamOptions,
	dir esdb.Direction,
	count uint64,
) (*esdb.ReadStream, error) {
	var fromPosition esdb.AllPosition
	switch options.Position {
	case store.Start:
		fromPosition = esdb.Start{}
	case store.End:
		fromPosition = esdb.End{}
	default:
		fromPosition = esdb.Position{Commit: uint64(options.Position)}
	}
	return es.client.ReadAll(ctx, esdb.ReadAllOptions{
		Direction:      dir,
		From:           fromPosition,
		ResolveLinkTos: false,
		Authenticated:  nil,
		Deadline:       nil,
		RequiresLeader: false,
	}, count)
}

func (es *EventStore) TruncateStream(ctx context.Context, streamID store.StreamID, opts ...store.TruncateStreamOption) error {
	//TODO implement me
	panic("implement me")
}

func (es *EventStore) DeleteStream(ctx context.Context, id store.StreamID) error {
	//TODO implement me
	panic("implement me")
}

func (es *EventStore) SubscribeToStream(ctx context.Context, streamID store.StreamID, opts ...store.SubscribeToStreamOption) (store.Subscription, error) {
	//TODO implement me
	panic("implement me")
}

func (es *EventStore) StreamExists(ctx context.Context, id store.StreamID) (bool, error) {
	ropts := esdb.ReadStreamOptions{
		From: esdb.Revision(10),
	}

	stream, err := es.client.ReadStream(ctx, string(id), ropts, 10)
	if err != nil {
		return false, errors.Wrapf(err, "failed checking if stream %s exists", id)
	}

	defer stream.Close()

	for {
		_, err := stream.Recv()
		if err, ok := esdb.FromError(err); !ok {
			if err.Code() == esdb.ErrorCodeResourceNotFound {
				return false, nil
			}
			return false, errors.Wrapf(err, "failed checking if stream %s exists", id)
		}
		return true, nil
	}
}

// GetStream this method is not supported by the esdb client.
func (es *EventStore) GetStream(ctx context.Context, id store.StreamID) (store.Stream, error) {
	exists, err := es.StreamExists(ctx, id)
	if err != nil {
		return store.Stream{}, errors.Wrapf(err, "failed getting information for stream %s", id)
	}
	if !exists {
		return store.Stream{}, store.StreamNotFoundError{StreamID: id}
	}

	panic("method not supported by client")
}

// Clear is not supported by the esdb client.
func (es *EventStore) Clear(ctx context.Context) error {
	panic("method not supported by client")
}
