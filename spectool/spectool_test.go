package spectool

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSpectool(t *testing.T) {
	specTool := Default("./spec_testdata/system.spec.yaml")
	err := specTool(context.Background())
	if err != nil {
		panic(err)
	}
	assert.NoError(t, err)
}
