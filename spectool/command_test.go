package spectool

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCommandFieldsMustHaveNameLinter(t *testing.T) {
	_, errs := CommandFieldsMustHaveNameLinter()(Spec{}, SpecGroup{
		Spec{
			Type:        CommandType,
			TypeName:    "",
			Description: "",
			Annotations: nil,
			Version:     "",
			Source:      SpecSource{},
			Properties: CommandSpecProperties{
				map[string]CommandField{
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

func TestCommandFieldsShouldHaveDescriptionLinter(t *testing.T) {
	warnings, _ := CommandFieldsShouldHaveDescriptionLinter()(Spec{}, SpecGroup{
		Spec{
			Type:        CommandType,
			TypeName:    "",
			Description: "",
			Annotations: nil,
			Version:     "",
			Source:      SpecSource{},
			Properties: CommandSpecProperties{
				map[string]CommandField{
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
