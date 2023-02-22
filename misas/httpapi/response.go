package httpapi

import (
	"github.com/morebec/go-errors/errors"
	"net/http"
)

// ResponseStatus represents the status of an API response.
type ResponseStatus string

const Success ResponseStatus = "success"
const Failure ResponseStatus = "failure"

// EndpointResponse represents a response from the API.
// If the response is sent following a successful request, the status will be set to "success" (Success) and it may contain
// data, and the error field will be null.
// If the response is sent following an unsuccessful request, the status will be set to "failure" (Failure), the data field will be empty
// and the error field will be filled with an Error.
type EndpointResponse struct {
	StatusCode int            `json:"-"`
	Headers    http.Header    `json:"-"`
	Status     ResponseStatus `json:"status"`
	Data       any            `json:"data"`
	Error      *Error         `json:"error"`
}

func (r EndpointResponse) WithHeader(header, value string) EndpointResponse {
	r.Headers.Clone().Set(header, value)
	return r
}

func (r EndpointResponse) WithStatusCode(c int) EndpointResponse {
	r.StatusCode = c
	return r
}

// Error Represents an error that occurred during the processing of a request.
type Error struct {
	// Represents the type name of the error that can easily be checked against by clients e.g. access.unauthorized, access.denied, server.failure
	Type string `json:"type"`
	// A human-readable description of the error to aid developers of clients to the API debug their implementation.
	Message string `json:"message"`
	// Additional data about the error such as a list of field names and their invalid values.
	Data any `json:"data"`
}

// NewSuccessResponse creates a new Successful API response.
func NewSuccessResponse(data any) EndpointResponse {
	return EndpointResponse{
		StatusCode: http.StatusOK,
		Status:     Success,
		Data:       data,
		Error:      nil,
	}
}

// NewFailureResponse creates a new Failed API response.
func NewFailureResponse(errorType string, message string, data any) EndpointResponse {
	return EndpointResponse{
		StatusCode: http.StatusConflict,
		Status:     Failure,
		Data:       nil,
		Error: &Error{
			Type:    errorType,
			Message: message,
			Data:    data,
		},
	}
}

// NewErrorResponse returns a new response based on an error, it tries to find if an error code can be inferred from the error, if not
// wil return an InternalErrorResponse.
func NewErrorResponse(err error) EndpointResponse {
	if errWithCode, ok := err.(interface {
		Code() string
		Error() string
	}); ok {
		errorCode := errWithCode.Code()
		r := NewFailureResponse(errorCode, errWithCode.Error(), nil)
		if errorCode == errors.NotFoundCode {
			r = r.WithStatusCode(http.StatusNotFound)
		}

		if errorCode == errors.InternalErrorCode {
			r = r.WithStatusCode(http.StatusInternalServerError)
		}

		return r
	}

	return NewInternalErrorResponse(err)
}

const InternalErrorResponseType = "internal_error"

// NewInternalErrorResponse creates a new Failed API response from an internal error resulting in an error 500.
func NewInternalErrorResponse(err error) EndpointResponse {
	return NewFailureResponse(InternalErrorResponseType, err.Error(), nil).WithStatusCode(http.StatusInternalServerError)
}
