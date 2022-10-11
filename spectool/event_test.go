package spectool

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEventFieldsShouldHaveDescriptionLinter(t *testing.T) {
	warnings, _ := EventFieldsShouldHaveDescriptionLinter()(Spec{}, SpecGroup{
		Spec{
			Type:        EventType,
			TypeName:    "",
			Description: "",
			Annotations: nil,
			Version:     "",
			Source:      SpecSource{},
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
