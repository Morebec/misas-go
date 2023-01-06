package errors

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestError_Code(t *testing.T) {
	err := New(NotFoundCode)
	assert.Equal(t, NotFoundCode, err.Code())
}

func TestError_Error(t *testing.T) {
	err := Error{
		Message: "hello world",
		code:    "",
		Cause:   nil,
	}

	assert.Equal(t, "hello world", err.Error())
}

func TestError_Unwrap(t *testing.T) {
	wrapped := errors.New("wrapped error")
	err := Error{
		Message: "hello world",
		code:    "",
		Cause:   wrapped,
	}

	assert.Equal(t, wrapped, err.Unwrap())
}

func TestNew(t *testing.T) {
	err := New(NotFoundCode)

	assert.Equal(t, NotFoundCode, err.Code())
}

func TestWrap(t *testing.T) {
	wrapped := errors.New("wrapped error")
	err := Wrap(wrapped, NotFoundCode)

	assert.Equal(t, wrapped, err.Unwrap())
}

func TestWrapWithMessage(t *testing.T) {
	wrapped := errors.New("wrapped error")
	err := WrapWithMessage(wrapped, NotFoundCode, "hello world")

	assert.Equal(t, wrapped, err.Unwrap())
	assert.Equal(t, "hello world", err.Error())
}

func TestNewWithMessage(t *testing.T) {
	err := NewWithMessage(NotFoundCode, "hello world")

	assert.Equal(t, NotFoundCode, err.Code())
	assert.Equal(t, "hello world", err.Error())
}

func TestHasCode(t *testing.T) {
	notFound := New(NotFoundCode)
	alreadyExists := New(AlreadyExists)

	assert.True(t, HasCode(notFound, NotFoundCode))
	assert.False(t, HasCode(alreadyExists, NotFoundCode))
	assert.False(t, HasCode(errors.New("not not found"), NotFoundCode))

	wrapping := Wrap(
		Wrap(
			errors.New("std error"),
			NotFoundCode,
		),
		AlreadyExists,
	)
	assert.True(t, HasCode(wrapping, AlreadyExists))
	assert.True(t, HasCode(wrapping, NotFoundCode))
	assert.False(t, HasCode(wrapping, "not code"))
}

func TestHasCodeStrict(t *testing.T) {
	notFound := New(NotFoundCode)
	alreadyExists := New(AlreadyExists)

	assert.True(t, HasCodeStrict(notFound, NotFoundCode))
	assert.False(t, HasCodeStrict(alreadyExists, NotFoundCode))
	assert.False(t, HasCodeStrict(errors.New("not not found"), NotFoundCode))

	wrapping := Wrap(
		Wrap(
			errors.New("std error"),
			NotFoundCode,
		),
		AlreadyExists,
	)
	assert.True(t, HasCodeStrict(wrapping, AlreadyExists))
	assert.False(t, HasCodeStrict(wrapping, NotFoundCode))
	assert.False(t, HasCodeStrict(wrapping, "not code"))
}

func Test_RootCause(t *testing.T) {
	stdErr := errors.New("std error")
	err := Wrap(
		Wrap(
			Wrap(
				stdErr,
				"standard",
			),
			NotFoundCode,
		),
		AlreadyExists,
	)
	assert.Equal(t, stdErr, RootCause(err))

	rc := New("no root cause")
	assert.Equal(t, rc, RootCause(rc))
}
