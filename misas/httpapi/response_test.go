package httpapi

import (
	"github.com/morebec/go-errors/errors"
	"net/http"
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
		expect   EndpointResponse
	}{
		{
			testName: "new internal error",
			given: args{
				err: errors.NewWithMessage(errors.NotFoundCode, "resource xyz not found"),
			},
			expect: EndpointResponse{
				StatusCode: http.StatusInternalServerError,
				Status:     Failure,
				Data:       nil,
				Error: &Error{
					Type:    InternalErrorResponseType,
					Message: "resource xyz not found",
					Data:    nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := NewInternalErrorResponse(tt.given.err); !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("NewInternalErrorResponse() = %v, expect %v", got, tt.expect)
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
		expect   EndpointResponse
	}{
		{
			testName: "new success response",
			given: args{
				data: 123,
			},
			expect: EndpointResponse{
				StatusCode: http.StatusOK,
				Status:     Success,
				Data:       123,
				Error:      nil,
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
		expect   EndpointResponse
	}{
		{
			testName: "",
			given: args{
				errorType: "unit-test-failed",
				message:   "the unit test failed",
				data:      123,
			},
			expect: EndpointResponse{
				StatusCode: http.StatusConflict,
				Status:     Failure,
				Data:       nil,
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
		expect   EndpointResponse
	}{
		{
			testName: "test new error response",
			given: GivenArgs{
				err: errors.NewWithMessage(errors.NotFoundCode, "resource xyz not found"),
			},
			expect: EndpointResponse{
				StatusCode: http.StatusNotFound,
				Status:     Failure,
				Data:       nil,
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
