package httpapi

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

func TestNewEndpointRequest(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		panic(err)
	}
	_, err = NewEndpointRequest(r, nil, AllowEmptyBody)
	assert.NoError(t, err)

	_, err = NewEndpointRequest(r, nil, DisallowEmptyBody)
	assert.NoError(t, err)
}

func TestEndpointRequest_Unmarshal(t *testing.T) {
	s := &struct {
		Hello string `json:"hello"`
	}{}
	r, err := http.NewRequest("GET", "/", strings.NewReader("{\"hello\": \"world\"}"))
	if err != nil {
		panic(err)
	}

	var er *EndpointRequest
	er, err = NewEndpointRequest(r, nil, AllowEmptyBody, DisallowUnknownFields)
	assert.NoError(t, err)
	err = er.Unmarshal(s)
	assert.NoError(t, err)

	r, err = http.NewRequest("GET", "/", strings.NewReader("{\"hello\": \"world\", \"unknown\": \"field\"}"))
	if err != nil {
		panic(err)
	}

	er, err = NewEndpointRequest(r, nil, AllowEmptyBody, DisallowUnknownFields)
	assert.NoError(t, err)
	err = er.Unmarshal(s)
	assert.Error(t, err)
}
