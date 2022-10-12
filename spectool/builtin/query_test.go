package builtin

import (
	"github.com/morebec/misas-go/spectool/spec"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQueryFieldsShouldHaveDescriptionLinter(t *testing.T) {
	warnings, _ := QueryFieldsShouldHaveDescriptionLinter()(spec.Spec{}, spec.Group{
		spec.Spec{
			Type:        QueryType,
			TypeName:    "",
			Description: "",
			Annotations: nil,
			Version:     "",
			Source:      spec.Source{},
			Properties: QuerySpecProperties{
				map[string]QueryField{
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

func TestQueryFieldsMustHaveTypeLinter(t *testing.T) {
	_, errs := QueryFieldsMustHaveTypeLinter()(spec.Spec{}, spec.Group{
		spec.Spec{
			Type:        QueryType,
			TypeName:    "",
			Description: "",
			Annotations: nil,
			Version:     "",
			Source:      spec.Source{},
			Properties: QuerySpecProperties{
				map[string]QueryField{
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
