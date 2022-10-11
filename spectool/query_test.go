package spectool

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQueryFieldsMustHaveNameLinter(t *testing.T) {
	_, errs := QueryFieldsMustHaveNameLinter()(Spec{}, SpecGroup{
		Spec{
			Type:        QueryType,
			TypeName:    "",
			Description: "",
			Annotations: nil,
			Version:     "",
			Source:      SpecSource{},
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

func TestQueryFieldsShouldHaveDescriptionLinter(t *testing.T) {
	warnings, _ := QueryFieldsShouldHaveDescriptionLinter()(Spec{}, SpecGroup{
		Spec{
			Type:        QueryType,
			TypeName:    "",
			Description: "",
			Annotations: nil,
			Version:     "",
			Source:      SpecSource{},
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
