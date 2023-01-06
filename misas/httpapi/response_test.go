package httpapi

import (
	"github.com/morebec/misas-go/misas/errors"
	"reflect"
	"testing"
)

func TestNewInternalError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		testName string
		given    args
		expect   Response
	}{
		{
			testName: "new internal error",
			given: args{
				err: errors.NewWithMessage(errors.NotFoundCode, "resource xyz not found"),
			},
			expect: Response{
				Status: Failure,
				Data:   nil,
				Error: &Error{
					Type:    InternalError,
					Message: "resource xyz not found",
					Data:    nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := NewInternalError(tt.given.err); !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("NewInternalError() = %v, expect %v", got, tt.expect)
			}
		})
	}
}

func TestNewSuccessResponse(t *testing.T) {
	type args struct {
		data any
	}
	tests := []struct {
		testName string
		given    args
		expect   Response
	}{
		{
			testName: "new success response",
			given: args{
				data: 123,
			},
			expect: Response{
				Status: Success,
				Data:   123,
				Error:  nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := NewSuccessResponse(tt.given.data); !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("NewSuccessResponse() = %v, expect %v", got, tt.expect)
			}
		})
	}
}

func TestNewFailureResponse(t *testing.T) {
	type args struct {
		errorType string
		message   string
		data      any
	}
	tests := []struct {
		testName string
		given    args
		expect   Response
	}{
		{
			testName: "",
			given: args{
				errorType: "unit-test-failed",
				message:   "the unit test failed",
				data:      123,
			},
			expect: Response{
				Status: Failure,
				Data:   nil,
				Error: &Error{
					Type:    "unit-test-failed",
					Message: "the unit test failed",
					Data:    123,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := NewFailureResponse(tt.given.errorType, tt.given.message, tt.given.data); !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("NewFailureResponse() = %v, expect %v", got, tt.expect)
			}
		})
	}
}

func TestNewErrorResponse(t *testing.T) {
	type GivenArgs struct {
		err error
	}
	tests := []struct {
		testName string
		given    GivenArgs
		expect   Response
	}{
		{
			testName: "test new error response",
			given: GivenArgs{
				err: errors.NewWithMessage(errors.NotFoundCode, "resource xyz not found"),
			},
			expect: Response{
				Status: Failure,
				Data:   nil,
				Error: &Error{
					Type:    errors.NotFoundCode,
					Message: "resource xyz not found",
					Data:    nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := NewErrorResponse(tt.given.err); !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("NewErrorResponse() = %v, expect %v", got, tt.expect)
			}
		})
	}
}
