package spectool

const SystemSpecType SpecType = "system"

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
	return func(systemSpec Spec, specs SpecGroup) ([]DependencyNode, error) {
		systemNode := NewDependencyNode(systemSpec, MapSpecGroup[SpecTypeName](specs, func(s Spec) SpecTypeName {
			return s.TypeName
		})...)

		nodes := []DependencyNode{systemNode}

		for _, spec := range specs {
			nodes = append(nodes, NewDependencyNode(spec))
		}

		return nodes, nil
	}
}
