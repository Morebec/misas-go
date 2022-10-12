package builtin

import (
	"github.com/morebec/misas-go/spectool/spec"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCommandFieldsShouldHaveDescriptionLinter(t *testing.T) {
	warnings, _ := CommandFieldsShouldHaveDescriptionLinter()(spec.Spec{}, spec.Group{
		spec.Spec{
			Type:        CommandType,
			TypeName:    "",
			Description: "",
			Annotations: nil,
			Version:     "",
			Source:      spec.Source{},
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

func TestCommandFieldsMustHaveTypeLinter(t *testing.T) {
	_, errs := CommandFieldsMustHaveTypeLinter()(spec.Spec{}, spec.Group{
		spec.Spec{
			Type:        CommandType,
			TypeName:    "",
			Description: "",
			Annotations: nil,
			Version:     "",
			Source:      spec.Source{},
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
