package postgresql

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func buildDocumentStore() *DocumentStore {
	ds := NewDocumentStore("postgres://postgres@localhost:5432/postgres?sslmode=disable")
	if err := ds.Open(context.Background()); err != nil {
		panic(err)
	}
	return ds
}

func TestDocumentStore_CreateCollection(t *testing.T) {
	ds := buildDocumentStore()
	if err := ds.CreateCollection(context.Background(), "test"); err != nil {
		panic(err)
	}

	err := ds.DeleteCollection(context.Background(), "test")
	assert.NoError(t, err)

	err = ds.DeleteCollection(context.Background(), "does-not-exist")
	assert.NoError(t, err)
}

func TestDocumentStore_DeleteCollection(t *testing.T) {
	ds := buildDocumentStore()
	if err := ds.CreateCollection(context.Background(), "test"); err != nil {
		panic(err)
	}
	err := ds.DeleteCollection(context.Background(), "test")
	assert.NoError(t, err)

	err = ds.DeleteCollection(context.Background(), "does-not-exist")
	assert.NoError(t, err)
}

func TestDocumentStore_InsertOne(t *testing.T) {
	type user struct {
		Id       string `json:"id"`
		Username string `json:"username"`
	}

	ds := buildDocumentStore()
	defer func(ds *DocumentStore, ctx context.Context, collectionName string) {
		_ = ds.DeleteCollection(ctx, collectionName)
	}(ds, context.Background(), "unit_test")

	u := user{
		Id:       "000",
		Username: "unit.test",
	}

	doc, err := NewDocument(u.Id, u)
	if err != nil {
		panic(err)
	}

	err = ds.InsertOne(context.Background(), "unit_test", doc)
	assert.NoError(t, err)

	// document already exists, should return error
	err = ds.InsertOne(context.Background(), "unit_test", doc)
	assert.Error(t, err)
}

func TestDocumentStore_InsertMany(t *testing.T) {
	type user struct {
		Id       string `json:"id"`
		Username string `json:"username"`
	}

	ds := buildDocumentStore()
	defer func(ds *DocumentStore, ctx context.Context, collectionName string) {
		_ = ds.DeleteCollection(ctx, collectionName)
	}(ds, context.Background(), "unit_test")
	var users []Document

	for i := 0; i < 3; i++ {
		doc, err := NewDocument(strconv.Itoa(i), user{
			Id:       strconv.Itoa(i),
			Username: fmt.Sprintf("user_%d", i),
		})
		if err != nil {
			panic(err)
		}
		users = append(users, doc)
	}

	err := ds.InsertMany(context.Background(), "unit_test", users)
	assert.NoError(t, err)

	// documents already exists, should return error
	err = ds.InsertMany(context.Background(), "unit_test", users)
	assert.Error(t, err)
}

func TestDocumentStore_UpsertOne(t *testing.T) {
	type user struct {
		Id       string `json:"id"`
		Username string `json:"username"`
	}

	ds := buildDocumentStore()
	defer func(ds *DocumentStore, ctx context.Context, collectionName string) {
		_ = ds.DeleteCollection(ctx, collectionName)
	}(ds, context.Background(), "unit_test")
	u := user{
		Id:       "000",
		Username: "unit.test",
	}

	doc, err := NewDocument(u.Id, u)
	if err != nil {
		panic(err)
	}

	err = ds.UpsertOne(context.Background(), "unit_test", doc)
	assert.NoError(t, err)

	u.Username = "upsert.test"
	doc, err = NewDocument(u.Id, u)
	if err != nil {
		panic(err)
	}
	err = ds.UpsertOne(context.Background(), "unit_test", doc)
	assert.NoError(t, err)
}

func TestDocumentStore_UpsertMany(t *testing.T) {
	type user struct {
		Id       string `json:"id"`
		Username string `json:"username"`
	}

	ds := buildDocumentStore()
	defer func(ds *DocumentStore, ctx context.Context, collectionName string) {
		_ = ds.DeleteCollection(ctx, collectionName)
	}(ds, context.Background(), "unit_test")

	var users []Document

	for i := 1; i < 3; i++ {
		doc, err := NewDocument(strconv.Itoa(i), user{
			Id:       strconv.Itoa(i),
			Username: fmt.Sprintf("user_%d", i),
		})
		if err != nil {
			panic(err)
		}
		users = append(users, doc)
	}

	err := ds.UpsertMany(context.Background(), "unit_test", users)
	assert.NoError(t, err)

	// documents already exists, should not return error
	err = ds.UpsertMany(context.Background(), "unit_test", users)
	assert.NoError(t, err)
}

func TestDocumentStore_UpdateOne(t *testing.T) {
	type user struct {
		Id       string `json:"id"`
		Username string `json:"username"`
	}

	ds := buildDocumentStore()
	defer func(ds *DocumentStore, ctx context.Context, collectionName string) {
		_ = ds.DeleteCollection(ctx, collectionName)
	}(ds, context.Background(), "unit_test")
	u := user{
		Id:       "000",
		Username: "unit.test",
	}

	doc, err := NewDocument(u.Id, u)
	if err != nil {
		panic(err)
	}

	err = ds.CreateCollection(context.Background(), "unit_test")
	assert.NoError(t, err)

	// SHOULD RETURN ERROR SINCE DOC NOT FOUND
	err = ds.UpdateOne(context.Background(), "unit_test", doc)
	assert.Error(t, err)

	// SHOULD BE UPDATED WHEN INSERTED FIRST
	err = ds.InsertOne(context.Background(), "unit_test", doc)
	assert.NoError(t, err)

	u.Username = "updated"
	doc, err = NewDocument(u.Id, u)

	err = ds.UpdateOne(context.Background(), "unit_test", doc)
	assert.NoError(t, err)
}

func TestDocumentStore_UpdateMany(t *testing.T) {
	type user struct {
		Id       string `json:"id"`
		Username string `json:"username"`
	}

	ds := buildDocumentStore()
	defer func(ds *DocumentStore, ctx context.Context, collectionName string) {
		_ = ds.DeleteCollection(ctx, collectionName)
	}(ds, context.Background(), "unit_test")

	var users []Document

	for i := 1; i < 3; i++ {
		doc, err := NewDocument(strconv.Itoa(i), user{
			Id:       strconv.Itoa(i),
			Username: fmt.Sprintf("user_%d", i),
		})
		if err != nil {
			panic(err)
		}
		users = append(users, doc)
	}

	err := ds.InsertMany(context.Background(), "unit_test", users)
	assert.NoError(t, err)

	users = nil
	for i := 1; i < 3; i++ {
		doc, err := NewDocument(strconv.Itoa(i), user{
			Id:       strconv.Itoa(i),
			Username: fmt.Sprintf("user_%d_updated", i),
		})
		if err != nil {
			panic(err)
		}
		users = append(users, doc)
	}

	// documents already exists, should not return error
	err = ds.UpdateMany(context.Background(), "unit_test", users)
	assert.NoError(t, err)
}

func TestDocumentStore_FindOneById(t *testing.T) {
	type user struct {
		Id       string `json:"id"`
		Username string `json:"username"`
	}

	ds := buildDocumentStore()
	defer func(ds *DocumentStore, ctx context.Context, collectionName string) {
		_ = ds.DeleteCollection(ctx, collectionName)
	}(ds, context.Background(), "unit_test")

	u := user{
		Id:       "000",
		Username: "unit.test",
	}

	doc, err := NewDocument(u.Id, u)
	if err != nil {
		panic(err)
	}
	err = ds.UpsertOne(context.Background(), "unit_test", doc)
	assert.NoError(t, err)

	found, err := ds.FindOneByID(context.Background(), "unit_test", u.Id)
	assert.NoError(t, err)

	var foundU user
	err = found.Unmarshall(&foundU)
	assert.NoError(t, err)
	assert.Equal(t, u, foundU)
}

func TestDocumentStore_FindOneBy(t *testing.T) {
	type user struct {
		Id       string `json:"id"`
		Username string `json:"username"`
	}

	ds := buildDocumentStore()
	defer func(ds *DocumentStore, ctx context.Context, collectionName string) {
		_ = ds.DeleteCollection(ctx, collectionName)
	}(ds, context.Background(), "unit_test")

	u := user{
		Id:       "000",
		Username: "unit.test",
	}

	doc, err := NewDocument(u.Id, u)
	if err != nil {
		panic(err)
	}
	err = ds.UpsertOne(context.Background(), "unit_test", doc)
	assert.NoError(t, err)

	found, err := ds.FindOneBy(context.Background(), "unit_test", "data ->> 'username' = $1", u.Username)
	assert.NoError(t, err)

	var foundU user
	err = found.Unmarshall(&foundU)
	assert.NoError(t, err)
	assert.Equal(t, u, foundU)
}

func TestDocumentStore_FindManyById(t *testing.T) {
	type user struct {
		Id       string `json:"id"`
		Username string `json:"username"`
	}

	ds := buildDocumentStore()
	defer func(ds *DocumentStore, ctx context.Context, collectionName string) {
		_ = ds.DeleteCollection(ctx, collectionName)
	}(ds, context.Background(), "unit_test")

	u := user{
		Id:       "000",
		Username: "unit.test",
	}

	doc, err := NewDocument(u.Id, u)
	if err != nil {
		panic(err)
	}
	err = ds.UpsertOne(context.Background(), "unit_test", doc)
	assert.NoError(t, err)

	found, err := ds.FindOneByID(context.Background(), "unit_test", u.Id)
	assert.NoError(t, err)

	var foundU user
	err = found.Unmarshall(&foundU)
	assert.NoError(t, err)
	assert.Equal(t, u, foundU)
}

func TestDocumentStore_FindManyBy(t *testing.T) {
	type user struct {
		Id       string `json:"id"`
		Username string `json:"username"`
		Enabled  bool   `json:"enabled"`
	}

	ds := buildDocumentStore()
	defer func(ds *DocumentStore, ctx context.Context, collectionName string) {
		_ = ds.DeleteCollection(ctx, collectionName)
	}(ds, context.Background(), "unit_test")

	var users []Document

	for i := 0; i < 3; i++ {
		doc, err := NewDocument(strconv.Itoa(i), user{
			Id:       strconv.Itoa(i),
			Username: fmt.Sprintf("user_%d", i),
			Enabled:  true,
		})
		if err != nil {
			panic(err)
		}
		users = append(users, doc)
	}

	err := ds.InsertMany(context.Background(), "unit_test", users)
	assert.NoError(t, err)

	// documents already exists, should return error
	err = ds.InsertMany(context.Background(), "unit_test", users)
	assert.Error(t, err)

	docs, err := ds.FindBy(context.Background(), "unit_test", "(data->>'enabled')::bool = $1", true)
	assert.NoError(t, err)

	assert.Len(t, docs, 3)
}

func TestCollection_Create(t *testing.T) {
	ds := buildDocumentStore()
	col := ds.Collection("unit.test")
	err := col.Create(context.Background())
	assert.NoError(t, err)
}

func TestDocumentStore_DeleteOneById(t *testing.T) {
	type user struct {
		Id       string `json:"id"`
		Username string `json:"username"`
	}

	ds := buildDocumentStore()

	u := user{
		Id:       "000",
		Username: "unit.test",
	}

	doc, err := NewDocument(u.Id, u)
	if err != nil {
		panic(err)
	}

	err = ds.InsertOne(context.Background(), "unit_test", doc)
	assert.NoError(t, err)

	// document already exists, should return error
	err = ds.InsertOne(context.Background(), "unit_test", doc)
	assert.Error(t, err)

	if err := ds.DeleteCollection(context.Background(), "unit_test"); err != nil {
		panic(err)
	}
}

func TestDocumentStore_DeleteBy(t *testing.T) {
	type user struct {
		Id       string `json:"id"`
		Username string `json:"username"`
		Enabled  bool   `json:"enabled"`
	}

	ds := buildDocumentStore()
	defer func(ds *DocumentStore, ctx context.Context, collectionName string) {
		_ = ds.DeleteCollection(ctx, collectionName)
	}(ds, context.Background(), "unit_test")

	var users []Document

	for i := 0; i < 3; i++ {
		doc, err := NewDocument(strconv.Itoa(i), user{
			Id:       strconv.Itoa(i),
			Username: fmt.Sprintf("user_%d", i),
			Enabled:  true,
		})
		if err != nil {
			panic(err)
		}
		users = append(users, doc)
	}

	err := ds.InsertMany(context.Background(), "unit_test", users)
	assert.NoError(t, err)

	err = ds.DeleteBy(context.Background(), "unit_test", "(data->>'enabled')::bool = $1", true)
	assert.NoError(t, err)

	docs, err := ds.FindBy(context.Background(), "unit_test", "(data->>'enabled')::bool = $1", true)
	assert.NoError(t, err)

	assert.Len(t, docs, 0)
}
