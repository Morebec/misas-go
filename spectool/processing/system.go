package processing

import (
	"github.com/morebec/misas-go/spectool/spec"
)

const SystemSpecType spec.Type = "system"

type SystemSpecProperties struct {
	License struct {
		Name string `yaml:"name"`
		URL  string `yaml:"url"`
	} `yaml:"license"`
	Contacts []struct {
		Name         string `yaml:"name"`
		EmailAddress string `yaml:"emailAddress"`
		URL          string `yaml:"url"`
	} `yaml:"contacts"`

	// Additional user defined locations for sources.
	Sources []string
}

func (s SystemSpecProperties) IsSpecProperties() {}

func SystemSpecDeserializer() SpecDeserializer {
	return NewTypePropertiesDeserializer[SystemSpecProperties](SystemSpecType)
}

func SystemDependencyProvider() DependencyProvider {
	return func(systemSpec spec.Spec, specs spec.Group) ([]DependencyNode, error) {
		systemNode := NewDependencyNode(systemSpec, spec.MapSpecGroup[spec.TypeName](specs, func(s spec.Spec) spec.TypeName {
			return s.TypeName
		})...)

		nodes := []DependencyNode{systemNode}

		for _, s := range specs {
			nodes = append(nodes, NewDependencyNode(s))
		}

		return nodes, nil
	}
}
