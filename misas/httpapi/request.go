package httpapi

import (
	"encoding/json"
	"fmt"
	"github.com/golang/gddo/httputil/header"
	"github.com/morebec/go-errors/errors"
	"io"
	"log"
	"net/http"
	"strings"
)

// EndpointRequest is a struct representing a request to an JSON API Endpoint.
// It wraps a request and provides helper methods to work with a JSON API.
type EndpointRequest struct {
	*http.Request
	responseWriter http.ResponseWriter
	// allows defining that unknown fields should cause an error when unmarshal
	DisallowUnknownFields bool
	DisallowEmptyBody     bool
	MaxBodyBytes          int64
}

type RequestOption func(r *EndpointRequest)

func DisallowUnknownFields(r *EndpointRequest) {
	r.DisallowUnknownFields = true
}

func AllowUnknownFields(r *EndpointRequest) {
	r.DisallowUnknownFields = false
}

func DisallowEmptyBody(r *EndpointRequest) {
	r.DisallowEmptyBody = true
}

func AllowEmptyBody(r *EndpointRequest) {
	r.DisallowEmptyBody = false
}

func MaxBodyBytes(value int64) RequestOption {
	return func(r *EndpointRequest) {
		r.MaxBodyBytes = value
	}
}

const BadRequestErrorCode = "invalid_request"

// NewEndpointRequest creates a new endpoint response instance for the provided request and validates that it is a correct
// an JSON API request.
// If no options are passed defaults to
// - A 1MB maximum body size
// - Disallowed Unknown fields
// - Disallowed empty body.
// Errors returned by this method can be directly passed to the NewErrorResponse without wrapping.
func NewEndpointRequest(r *http.Request, w http.ResponseWriter, opts ...RequestOption) (*EndpointRequest, error) {
	er := &EndpointRequest{Request: r, responseWriter: w}

	// Set default options.
	opts = append([]RequestOption{
		MaxBodyBytes(1048576), // 1MB
		DisallowUnknownFields,
		DisallowEmptyBody,
	}, opts...)

	for _, opt := range opts {
		opt(er)
	}

	if err := er.validateRequest(); err != nil {
		return nil, errors.WrapWithMessage(err, BadRequestErrorCode, "invalid r")
	}

	if er.MaxBodyBytes != 0 {
		er.Body = http.MaxBytesReader(er.responseWriter, er.Body, er.MaxBodyBytes)
	}

	return er, nil
}

// Unmarshal a JSON body to a provided value.
// Errors returned by this method can be directly passed to the NewErrorResponse without wrapping.
func (r *EndpointRequest) Unmarshal(v any) error {
	decoder := json.NewDecoder(r.Body)
	if r.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	err := decoder.Decode(v)

	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		// Catch common unmarshal errors and provide appropriate error codes and messages.
		// We do this to obfuscate the fact that we are using Go to clients, and ensure they have relevant error messages.
		switch {
		// Catch JSON syntax errors.
		case errors.As(err, &syntaxError):
			return errors.NewWithMessage(
				BadRequestErrorCode,
				fmt.Sprintf("EndpointRequest body contains malformed json at position %d", syntaxError.Offset),
			)

			// Catch JSON syntax errors.
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.NewWithMessage(
				BadRequestErrorCode,
				"EndpointRequest body contains malformed json",
			)
			// Catch unmarshal errors, such as incompatible types (e.g. string in request body to int in struct field).
		case errors.As(err, &unmarshalTypeError):
			return errors.NewWithMessage(
				BadRequestErrorCode,
				fmt.Sprintf(
					"invalid value for field %q at position %d",
					unmarshalTypeError.Field,
					unmarshalTypeError.Offset,
				),
			)

			// Catch errors caused by extra fields when DisallowUnknownFields is enabled.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return errors.NewWithMessage(
				BadRequestErrorCode,
				fmt.Sprintf("unexpected field in request body: %s", fieldName),
			)

			// Catch in case of empty request body as in ""
		case errors.Is(err, io.EOF):
			if r.DisallowEmptyBody {
				return errors.NewWithMessage(BadRequestErrorCode, "request body must not be empty")
			}
			// Catch body too large
		case err.Error() == "http: request body too large":
			return errors.NewWithMessage(BadRequestErrorCode, "request body too larger")

			// Catch all other errors.
		default:
			log.Print(err.Error())
			return errors.NewWithMessage(errors.InternalErrorCode, "unknown error")
		}
	}

	return nil
}

// GetQueryParam returns the URL Query Param of a given name. It acts as a shorthand to http.Request.URL.Query().Get(param)
func (r *EndpointRequest) GetQueryParam(param string) string {
	return r.URL.Query().Get(param)
}

func (r *EndpointRequest) GetQueryParamOrDefault(param string, v string) string {
	if !r.URL.Query().Has(param) {
		return v
	}
	return r.URL.Query().Get(param)
}

// internal method that validates that the request is indeed a valid JSON API request.
func (r *EndpointRequest) validateRequest() error {
	// If we have a content-type, ensure it is application/json,
	// we use the github.com/golang/gddo/httputil/header library to ensure that if the content type contains
	// additional charset parameters, that we correctly parse it.
	if contentType, _ := header.ParseValueAndParams(r.Header, "Content-Type"); contentType != "" {
		if contentType != "application/json" {
			return errors.NewWithMessage(
				BadRequestErrorCode,
				fmt.Sprintf("expected content-type header to be application/json got %s", contentType),
			)
		}
	}

	return nil
}
