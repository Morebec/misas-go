package builtin

import (
	"github.com/morebec/misas-go/spectool/spec"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEventFieldsShouldHaveDescriptionLinter(t *testing.T) {
	warnings, _ := EventFieldsShouldHaveDescriptionLinter()(spec.Spec{}, spec.Group{
		spec.Spec{
			Type:        EventType,
			TypeName:    "",
			Description: "",
			Annotations: nil,
			Version:     "",
			Source:      spec.Source{},
			Properties: EventSpecProperties{
				map[string]EventField{
					"field": {
						Name:        "",
						Description: "",
						Type:        "",
						Nullable:    false,
						Deprecation: "",
						Example:     "",
						Annotations: nil,
					},
				},
			},
		},
	})

	assert.NotEmpty(t, warnings)
}
