package event

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInMemoryBus_RegisterHandler(t *testing.T) {
	b := NewInMemoryBus()

	b.RegisterHandler(unitTestSucceededTypeName, HandlerFunc(func(ctx context.Context, e Event) error {
		return nil
	}))
}

func TestInMemoryBus_Send(t *testing.T) {
	b := NewInMemoryBus()

	sent := false
	b.RegisterHandler(unitTestFailedTypeName, HandlerFunc(func(ctx context.Context, e Event) error {
		sent = true
		return nil
	}))

	err := b.Send(context.Background(), New(unitTestFailed{}))
	assert.NoError(t, err)

	assert.True(t, sent)
}

func TestNewInMemoryBus(t *testing.T) {
	assert.NotNil(t, NewInMemoryBus())
}
