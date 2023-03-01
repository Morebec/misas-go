package httpapi

import (
	"testing"
)

func TestWebServer_Start(t *testing.T) {
	ws := NewServer()

	if err := ws.Start(); err != nil {
		panic(err)
	}
}
