package domain

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVersion_Incremented(t *testing.T) {
	v := Version(5)
	assert.Equal(t, Version(6), v.Incremented())
}
