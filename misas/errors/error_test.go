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

func TestGroup_Code(t *testing.T) {
	g := Group{
		code: InternalErrorCode,
	}

	assert.Equal(t, InternalErrorCode, g.Code())
}

func TestNewGroup(t *testing.T) {
	g := NewGroup(
		InternalErrorCode,
		New(NotFoundCode),
		New(AlreadyExists),
	)

	assert.Equal(t, Group{
		code: InternalErrorCode,
		Errors: []error{
			New(NotFoundCode),
			New(AlreadyExists),
		},
		Cause: nil,
	}, g)
}

func TestWrapAsGroup(t *testing.T) {
	type args struct {
		code  string
		cause error
		errs  []error
	}

	stdErr := errors.New("stderror")
	err1 := New("error_1")
	err2 := New("error_2")

	tests := []struct {
		testName string
		given    args
		expect   Group
	}{
		{
			testName: "wrap with a group",
			given: args{
				code:  InternalErrorCode,
				cause: stdErr,
				errs: []error{
					err1,
					err2,
				},
			},
			expect: Group{
				code:  InternalErrorCode,
				Cause: stdErr,
				Errors: []error{
					err1,
					err2,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			assert.Equalf(t, tt.expect, WrapAsGroup(tt.given.code, tt.given.cause, tt.given.errs...), "WrapAsGroup(%v, %v, %v)", tt.given.code, tt.given.cause, tt.given.errs)
		})
	}
}

func TestGroup_WithError(t *testing.T) {
	errGrp := NewGroup(InternalErrorCode, New(NotFoundCode))
	errGrp = errGrp.WithError(New(NotFoundCode))

	assert.Len(t, errGrp.Errors, 2)
}

func TestGroup_ErrorS(t *testing.T) {
	errGrp := NewGroup(InternalErrorCode)
	assert.Equal(t, "no errors", errGrp.Error())

	errGrp = errGrp.WithError(New(NotFoundCode))
	assert.Equal(t, "not_found", errGrp.Error())

	errGrp = errGrp.WithError(New(NotFoundCode))
	assert.Equal(t, "not_found, and 1 other error(s)", errGrp.Error())
}

func TestGroup_Error(t *testing.T) {
	type fields struct {
		code   string
		Errors []error
		Cause  error
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "without any errors",
			fields: fields{
				code:   InternalErrorCode,
				Errors: nil,
				Cause:  nil,
			},
			want: "no errors",
		},
		{
			name: "with 1 error",
			fields: fields{
				code: InternalErrorCode,
				Errors: []error{
					NewWithMessage(NotFoundCode, "resource xyz was not found"),
				},
				Cause: nil,
			},
			want: "resource xyz was not found",
		},
		{
			name: "with two or more errors",
			fields: fields{
				code: InternalErrorCode,
				Errors: []error{
					NewWithMessage(NotFoundCode, "resource xyz was not found"),
					NewWithMessage(NotFoundCode, "resource abc was not found"),
				},
				Cause: nil,
			},
			want: "resource xyz was not found, and 1 other error(s)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := Group{
				code:   tt.fields.code,
				Errors: tt.fields.Errors,
				Cause:  tt.fields.Cause,
			}
			assert.Equalf(t, tt.want, g.Error(), "Error()")
		})
	}
}

func TestError_Error1(t *testing.T) {
	type GivenArgs struct {
		Message string
		code    string
		Cause   error
	}
	tests := []struct {
		testName string
		given    GivenArgs
		expect   string
	}{
		{
			testName: "with message should return message",
			given: GivenArgs{
				Message: "there was an error",
				code:    "",
				Cause:   nil,
			},
			expect: "there was an error",
		},
		{
			testName: "without message should return code",
			given: GivenArgs{
				Message: "",
				code:    "internal error",
				Cause:   nil,
			},
			expect: "internal error",
		},
		{
			testName: "with cause should also return the wrapped error message",
			given: GivenArgs{
				Message: "failed retrieving resource xyz",
				code:    "internal error",
				Cause:   NewWithMessage(InternalErrorCode, "database connection could not be established"),
			},
			expect: "failed retrieving resource xyz: database connection could not be established",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			e := Error{
				Message: tt.given.Message,
				code:    tt.given.code,
				Cause:   tt.given.Cause,
			}
			assert.Equalf(t, tt.expect, e.Error(), "Error()")
		})
	}
}
