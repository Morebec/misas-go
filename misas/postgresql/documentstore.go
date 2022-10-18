package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
)

// Document represents a document to be stored in the DocumentStore.
type Document struct {
	id   string
	data json.RawMessage
}

// NewDocument creates a new Document from a value and marshall it to json.
func NewDocument(id string, v any) (Document, error) {
	d := Document{
		id: id,
	}
	err := d.marshall(v)
	if err != nil {
		return Document{}, err
	}

	return d, nil
}

func (d *Document) marshall(v any) error {
	bytes, err := json.Marshal(v)
	if err != nil {
		return errors.Wrapf(err, "failed encoding document")
	}

	d.data = bytes

	return nil
}

// RecordedDocument represents a document that was stored inside the DocumentStore.
type RecordedDocument struct {
	ID   string
	data json.RawMessage
}

// Unmarshall the document to a value.
func (d RecordedDocument) Unmarshall(v any) error {
	if err := json.Unmarshal(d.data, v); err != nil {
		return errors.Wrapf(err, "failed decoding document")
	}

	return nil
}

// DocumentStore is an implementation of a simple document store using PostgreSQL.
// It creates a table for every collection.
type DocumentStore struct {
	connectionString string
	conn             *sql.DB
}

func NewDocumentStore(connectionString string) *DocumentStore {
	return &DocumentStore{connectionString: connectionString}
}

// Open a connection to the DocumentStore.
func (ds *DocumentStore) Open(ctx context.Context) error {
	operationFailed := func(err error) error {
		return errors.Wrap(err, "failed opening connection to document store")
	}
	conn, err := sql.Open("postgres", ds.connectionString)
	if err != nil {
		return operationFailed(err)
	}
	ds.conn = conn

	if err = ds.conn.PingContext(ctx); err != nil {
		return operationFailed(err)
	}

	if err := ds.setupSchema(ctx); err != nil {
		return operationFailed(err)
	}

	return nil
}

// Close the connection to the DocumentStore.
func (ds *DocumentStore) Close() error {
	if err := ds.conn.Close(); err != nil {
		return errors.Wrap(err, "failed closing connection to document store")
	}
	return nil
}

// Connection returns the connection pool to the DocumentStore.
func (ds *DocumentStore) Connection() *sql.DB {
	return ds.conn
}

// BeginTransaction begins a transaction.
func (ds *DocumentStore) BeginTransaction(ctx context.Context) (*sql.Tx, error) {
	return ds.conn.BeginTx(ctx, nil)
}

// Collection returns a Collection object that acts as a scoped proxy dor the DocumentStore where operations apply to the collection.
func (ds *DocumentStore) Collection(name string) Collection {
	return Collection{
		name: name,
		ds:   ds,
	}
}

// CreateCollection creates a new collection in the document store.
func (ds *DocumentStore) CreateCollection(ctx context.Context, collectionName string) error {
	if collectionName == "" {
		return errors.New("cannot create a collection named \"\"")
	}

	// begin transaction
	tx, err := ds.conn.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrapf(err, "failed creating collection %s", collectionName)
	}

	// Create Collection Table
	createCollectionTableSql := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS "%s" (
    id   VARCHAR(255) NOT NULL PRIMARY KEY,
	data JSONB
);`, collectionName)
	if _, err := ds.conn.ExecContext(ctx, createCollectionTableSql); err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.Wrapf(err, "failed creating collection %s", collectionName)
		}
		return errors.Wrapf(err, "failed creating collection %s", collectionName)
	}

	// Add to list of collections.
	if _, err := ds.conn.ExecContext(
		ctx,
		"INSERT INTO document_store_collections (collection_name) VALUES ($1) ON CONFLICT DO NOTHING",
		collectionName,
	); err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.Wrapf(err, "failed creating collection %s", collectionName)
		}
		return errors.Wrapf(err, "failed creating collection %s", collectionName)
	}

	// commit transaction
	if err := tx.Commit(); err != nil {
		return errors.Wrapf(err, "failed creating collection %s", collectionName)
	}

	return nil
}

// DeleteCollection deletes a collection from the document store.
func (ds *DocumentStore) DeleteCollection(ctx context.Context, collectionName string) error {
	if collectionName == "" {
		return errors.New("cannot delete a collection named \"\"")
	}

	// begin transaction
	tx, err := ds.conn.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrapf(err, "failed creating collection %s", collectionName)
	}

	// Create Collection Table
	createCollectionTableSql := fmt.Sprintf(`DROP TABLE IF EXISTS "%s";`, collectionName)
	if _, err := ds.conn.ExecContext(ctx, createCollectionTableSql); err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.Wrapf(err, "failed creating collection %s", collectionName)
		}
		return errors.Wrapf(err, "failed creating collection %s", collectionName)
	}

	// Add to list of collections.
	if _, err := ds.conn.ExecContext(
		ctx,
		"DELETE FROM document_store_collections WHERE collection_name = $1",
		collectionName,
	); err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.Wrapf(err, "failed creating collection %s", collectionName)
		}
		return errors.Wrapf(err, "failed creating collection %s", collectionName)
	}

	// commit transaction
	if err := tx.Commit(); err != nil {
		return errors.Wrapf(err, "failed creating collection %s", collectionName)
	}

	return nil
}

// InsertOne document into a collection.
// If the collection does not exist, it will be created. if a document with the provided documentId already exists, will return an error.
func (ds *DocumentStore) InsertOne(ctx context.Context, collectionName string, d Document) error {
	if err := ds.CreateCollection(ctx, collectionName); err != nil {
		return errors.Wrapf(err, "failed inserting document into collection %s", collectionName)
	}

	insertQuery := fmt.Sprintf(`INSERT INTO "%s" (id, data) VALUES ($1, $2)`, collectionName)
	if _, err := ds.conn.ExecContext(ctx, insertQuery, d.id, d.data); err != nil {
		return errors.Wrapf(err, "failed inserting document into collection %s", collectionName)
	}

	return nil
}

// InsertMany documents in a collection.
func (ds *DocumentStore) InsertMany(ctx context.Context, collectionName string, docs []Document) error {
	tx, err := ds.BeginTransaction(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed inserting documents")
	}

	for _, d := range docs {
		if err := ds.InsertOne(ctx, collectionName, d); err != nil {
			if err := tx.Rollback(); err != nil {
				return errors.Wrapf(err, "failed inserting documents")
			}
			return errors.Wrapf(err, "failed inserting documents")
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrapf(err, "failed inserting documents")
	}

	return nil
}

// UpsertOne a document into a collection.
func (ds *DocumentStore) UpsertOne(ctx context.Context, collectionName string, d Document) error {
	if err := ds.CreateCollection(ctx, collectionName); err != nil {
		return errors.Wrapf(err, "failed upserting document into collection %s", collectionName)
	}

	upsertQuery := fmt.Sprintf(`
INSERT INTO "%s" (id, data) 
VALUES ($1, $2) 
ON CONFLICT (id) DO UPDATE
SET data = $2
`, collectionName)
	if _, err := ds.conn.ExecContext(ctx, upsertQuery, d.id, d.data); err != nil {
		return errors.Wrapf(err, "failed upserting document into collection %s", collectionName)
	}

	return nil
}

// UpsertMany documents into a collection.
func (ds *DocumentStore) UpsertMany(ctx context.Context, collectionName string, docs []Document) error {
	tx, err := ds.BeginTransaction(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed upserting documents")
	}

	for _, d := range docs {
		if err := ds.UpsertOne(ctx, collectionName, d); err != nil {
			if err := tx.Rollback(); err != nil {
				return errors.Wrapf(err, "failed upserting documents")
			}
			return errors.Wrapf(err, "failed upserting documents")
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrapf(err, "failed upserting documents")
	}

	return nil
}

// UpdateOne document of a given collection.
func (ds *DocumentStore) UpdateOne(ctx context.Context, collectionName string, d Document) error {
	upsertQuery := fmt.Sprintf(`
UPDATE "%s" 
SET data = $1 
WHERE id = $2
`, collectionName)
	updated, err := ds.conn.ExecContext(ctx, upsertQuery, d.data, d.id)
	if err != nil {
		return errors.Wrapf(err, "failed updating document %s in collection %s", d.id, collectionName)
	}

	rowsAffected, err := updated.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, "failed updating document %s in collection %s", d.id, collectionName)
	}

	if rowsAffected != 1 {
		return errors.Errorf("failed updating document %s in collection %s, document not found", d.id, collectionName)
	}

	return nil
}

// UpdateMany documents of a collection.
func (ds *DocumentStore) UpdateMany(ctx context.Context, collectionName string, docs []Document) error {
	tx, err := ds.BeginTransaction(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed updating documents")
	}

	for _, d := range docs {
		if err := ds.UpdateOne(ctx, collectionName, d); err != nil {
			if err := tx.Rollback(); err != nil {
				return errors.Wrapf(err, "failed updating documents")
			}
			return errors.Wrapf(err, "failed updating documents")
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrapf(err, "failed updating documents")
	}

	return nil
}

// FindOneByID returns a document by its ID.
func (ds *DocumentStore) FindOneByID(ctx context.Context, collectionName string, documentID string) (doc RecordedDocument, err error) {
	return ds.FindOneBy(ctx, collectionName, "id = $1", documentID)
}

// FindOneBy returns the first document matching a certain query.
func (ds *DocumentStore) FindOneBy(ctx context.Context, collectionName string, query string, args ...any) (doc RecordedDocument, err error) {
	rows, err := ds.conn.QueryContext(ctx, fmt.Sprintf(`SELECT id, data FROM "%s" WHERE %s`, collectionName, query), args...)
	defer func(rows *sql.Rows) {
		if closeErr := rows.Close(); closeErr != nil {
			err = errors.Wrapf(err, "failed finding document")
		}
	}(rows)
	if err != nil {
		return RecordedDocument{}, errors.Wrapf(err, "failed finding document")
	}

	docs, err := ds.processRows(rows)
	if err != nil {
		return RecordedDocument{}, errors.Wrapf(err, "failed finding document")
	}

	if len(docs) == 0 {
		// TODO Not found error
		return RecordedDocument{}, errors.Wrapf(err, "failed finding document")
	}

	return docs[0], nil
}

// FindBy returns documents matching a certain query.
func (ds *DocumentStore) FindBy(ctx context.Context, collectionName string, query string, args ...any) (documents []RecordedDocument, err error) {
	rows, err := ds.conn.QueryContext(ctx, fmt.Sprintf(`SELECT id, data FROM "%s" WHERE %s`, collectionName, query), args...)
	defer func(rows *sql.Rows) {
		if closeErr := rows.Close(); closeErr != nil {
			err = errors.Wrapf(err, "failed finding documents")
		}
	}(rows)
	if err != nil {
		return nil, errors.Wrap(err, "failed finding documents")
	}

	documents, err = ds.processRows(rows)
	if err != nil {

		return nil, errors.Wrap(err, "failed finding documents")
	}

	return documents, nil
}

// DeleteOneByID deletes a document from a collection by its ID
func (ds *DocumentStore) DeleteOneByID(ctx context.Context, collectionName string, documentID string) error {
	return ds.DeleteBy(ctx, collectionName, "id = $1", documentID)
}

// DeleteBy deletes documents from a collection by a certain query.
func (ds *DocumentStore) DeleteBy(ctx context.Context, collectionName string, query string, args ...any) error {
	if collectionName == "" {
		return errors.New("cannot delete from a collection named \"\"")
	}

	if _, err := ds.conn.ExecContext(ctx, fmt.Sprintf(`DELETE FROM "%s" WHERE %s`, collectionName, query), args...); err != nil {
		return errors.Wrapf(err, "failed deleting document from collection %s", collectionName)
	}

	return nil
}

func (ds *DocumentStore) processRows(rows *sql.Rows) ([]RecordedDocument, error) {
	var docs []RecordedDocument
	for rows.Next() {
		var data []byte
		var id string
		if err := rows.Scan(&id, &data); err != nil {
			return nil, errors.Wrap(err, "failed processing row")
		}
		docs = append(docs, RecordedDocument{
			ID:   id,
			data: data,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "failed processing rows")
	}

	return docs, nil
}

func (ds *DocumentStore) setupSchema(ctx context.Context) error {
	operationFailed := func(err error) error {
		return errors.Wrapf(err, "failed creating document_store_collections table")
	}

	createTableSql := `
CREATE TABLE IF NOT EXISTS document_store_collections
(
    collection_name VARCHAR(255) NOT NULL PRIMARY KEY
);
`

	if _, err := ds.conn.ExecContext(ctx, createTableSql); err != nil {
		return operationFailed(err)
	}

	return nil
}

type Collection struct {
	name string
	ds   *DocumentStore
}

func (c Collection) Create(ctx context.Context) error {
	return c.ds.CreateCollection(ctx, c.name)
}

func (c Collection) Delete(ctx context.Context) error {
	return c.ds.DeleteCollection(ctx, c.name)
}

func (c Collection) InsertOne(ctx context.Context, d Document) error {
	return c.ds.InsertOne(ctx, c.name, d)
}

func (c Collection) InsertMany(ctx context.Context, docs []Document) error {
	return c.ds.InsertMany(ctx, c.name, docs)
}

func (c Collection) UpsertOne(ctx context.Context, d Document) error {
	return c.ds.UpsertOne(ctx, c.name, d)
}

func (c Collection) UpsertMany(ctx context.Context, docs []Document) error {
	return c.ds.UpsertMany(ctx, c.name, docs)
}

func (c Collection) UpdateOne(ctx context.Context, d Document) error {
	return c.ds.UpdateOne(ctx, c.name, d)
}

func (c Collection) UpdateMany(ctx context.Context, docs []Document) error {
	return c.ds.UpdateMany(ctx, c.name, docs)
}

func (c Collection) FindOneByID(ctx context.Context, documentID string) (doc RecordedDocument, err error) {
	return c.ds.FindOneByID(ctx, c.name, documentID)
}

func (c Collection) FindOneBy(ctx context.Context, query string, args ...any) (doc RecordedDocument, err error) {
	return c.ds.FindOneBy(ctx, c.name, query, args)
}

func (c Collection) FindBy(ctx context.Context, query string, args ...any) (documents []RecordedDocument, err error) {
	return c.ds.FindBy(ctx, c.name, query, args)
}

func (c Collection) DeleteOneByID(ctx context.Context, documentID string) error {
	return c.ds.DeleteOneByID(ctx, c.name, documentID)
}

func (c Collection) DeleteBy(ctx context.Context, query string, args ...any) error {
	return c.ds.DeleteBy(ctx, c.name, query, args)
}
