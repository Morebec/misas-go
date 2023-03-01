package httpapi

import (
	"testing"
)

func TestWebServer_Start(t *testing.T) {
	ws := NewWebServer()

	if err := ws.Start(); err != nil {
		panic(err)
	}
}
