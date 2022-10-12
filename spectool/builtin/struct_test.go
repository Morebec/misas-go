package builtin

import (
	"github.com/morebec/misas-go/spectool/spec"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStructFieldsShouldHaveDescriptionLinter(t *testing.T) {
	warnings, _ := StructFieldsShouldHaveDescriptionLinter()(spec.Spec{}, spec.Group{
		spec.Spec{
			Type:        StructType,
			TypeName:    "",
			Description: "",
			Annotations: nil,
			Version:     "",
			Source:      spec.Source{},
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

func TestStructFieldsMustHaveTypeLinter(t *testing.T) {
	_, errs := StructFieldsMustHaveTypeLinter()(spec.Spec{}, spec.Group{
		spec.Spec{
			Type:        StructType,
			TypeName:    "",
			Description: "",
			Annotations: nil,
			Version:     "",
			Source:      spec.Source{},
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

	assert.NotEmpty(t, errs)
}
