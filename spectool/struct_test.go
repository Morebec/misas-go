package spectool

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStructFieldsShouldHaveDescriptionLinter(t *testing.T) {
	warnings, _ := StructFieldsShouldHaveDescriptionLinter()(Spec{}, SpecGroup{
		Spec{
			Type:        StructType,
			TypeName:    "",
			Description: "",
			Annotations: nil,
			Version:     "",
			Source:      SpecSource{},
			Properties: StructSpecProperties{
				map[string]StructField{
					"field": {
						Name:        "",
						Description: "",
						Type:        "",
						Nullable:    false,
						Deprecation: "",
						Example:     "",
						Default:     "",
						Required:    false,
						Annotations: nil,
					},
				},
			},
		},
	})

	assert.NotEmpty(t, warnings)
}
