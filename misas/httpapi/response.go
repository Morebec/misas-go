package httpapi

// ResponseStatus represents the status of an API response.
type ResponseStatus string

const Success ResponseStatus = "success"
const Failure ResponseStatus = "failure"

// Response represents a response from the API.
// If the response is sent following a successful request, the status will be set to "success" (Success) and it may contain
// data, and the error field will be null.
// If the response is sent following an unsuccessful request, the status will be set to "failure" (Failure), the data field will be empty
// and the error field will be filled with an Error.
type Response struct {
	Status ResponseStatus `json:"status"`
	Data   any            `json:"data"`
	Error  *Error         `json:"error"`
}

// Error Represents an error that occurred during the processing of a request.
type Error struct {
	// type testName that can easily be checked against by clients e.g. access.unauthorized, access.denied, server.failure
	Type string `json:"type"`
	// A human-readable description of the error, to aid developers of clients to the API debug their implementation.
	Message string `json:"message"`
	// Additional data about the error such as field testName and invalid values.
	Data any `json:"data"`
}

// NewSuccessResponse creates a new Successful API response.
func NewSuccessResponse(data any) Response {
	return Response{
		Status: Success,
		Data:   data,
		Error:  nil,
	}
}

// NewFailureResponse creates a new Failed API response.
func NewFailureResponse(errorType string, message string, data any) Response {
	return Response{
		Status: Failure,
		Data:   nil,
		Error: &Error{
			Type:    errorType,
			Message: message,
			Data:    data,
		},
	}
}

// NewErrorResponse returns a new response based on an error, it tries to find if an error code can be inferred from the error.
func NewErrorResponse(err error) Response {
	if errWithCode, ok := err.(interface {
		Code() string
		Error() string
	}); ok {
		return NewFailureResponse(errWithCode.Code(), errWithCode.Error(), nil)
	}

	return NewInternalError(err)
}

const InternalError = "internal_error"

// NewInternalError creates a new Failed API response from an internal error resulting in an error 500.
func NewInternalError(err error) Response {
	return Response{
		Status: Failure,
		Data:   nil,
		Error: &Error{
			Type:    InternalError,
			Message: err.Error(),
			Data:    nil,
		},
	}
}
