// Copyright 2022 Morébec
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package httpapi

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"net/http"
	"time"
)

const InternalError = "internal_error"

// Endpoint represents an endpoint.
type Endpoint func(r chi.Router)

func HomeEndpoint(r chi.Router) {
	r.Get("/", func(writer http.ResponseWriter, request *http.Request) {
		request.
			writer.WriteHeader(500)
		render.JSON(writer, request, NewSuccessResponse(nil))
	})
}

func main() error {
	r := chi.NewRouter()
	r.Use(middleware.AllowContentType("application/json"))
	r.Use(middleware.CleanPath)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(time.Second * 60))
	r.Use(render.SetContentType(render.ContentTypeJSON))

	HomeEndpoint(r)

	if err := http.ListenAndServe(":3000", r); err != nil {
		return err
	}
	return nil
}

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
	// type name that can easily be checked against by clients e.g. access.unauthorized, access.denied, server.failure
	Type string `json:"type"`
	// A human-readable description of the error, to aid developers of clients to the API debug their implementation.
	Message string `json:"message"`
	// Additional data about the error such as field name and invalid values.
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

// NewErrorResponse creates a new Failed API response.
func NewErrorResponse(errorType string, message string, data any) Response {
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

// NewInternalError creates a new Failed API response from an internal error.
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
