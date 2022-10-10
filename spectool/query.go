package spectool

const QueryType SpecType = "query"

type QueryField struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Type        DataType `yaml:"type"`
	Nullable    bool     `yaml:"nullable"`
	Deprecation string   `yaml:"deprecation"`
	Example     string   `yaml:"example"`
	Default     string   `yaml:"default"`
	Required    bool     `yaml:"required"`

	// Annotations are used to tag a field with specific data to indicate additional information about the field.
	// One useful tag is the personal_data tag that indicates that this field contains personal information.
	Annotations Annotations `yaml:"annotations"`
}
type QuerySpecProperties struct {
	Fields map[string]QueryField `yaml:"fields"`
}

func (c QuerySpecProperties) IsSpecProperties() {}

func QueryDeserializer() SpecDeserializer {
	inner := NewTypePropertiesDeserializer[QuerySpecProperties](QueryType)
	return NewTypeBasedDeserializer(QueryType, func(source SpecSource) (Spec, error) {
		spec, err := inner.Deserialize(source)
		if err != nil {
			return Spec{}, err
		}

		props := spec.Properties.(QuerySpecProperties)
		for fn, f := range props.Fields {
			f.Name = fn
			props.Fields[fn] = f
		}

		return spec, nil
	})
}

func QueryDependencyProvider() DependencyProvider {
	return func(systemSpec Spec, specs SpecGroup) ([]DependencyNode, error) {
		queries := specs.SelectType(QueryType)

		var nodes []DependencyNode

		for _, q := range queries {
			props := q.Properties.(QuerySpecProperties)
			var deps []SpecTypeName
			for _, f := range props.Fields {
				userDefined := f.Type.ExtractUserDefined()
				if userDefined != "" {
					deps = append(deps, SpecTypeName(userDefined))
				}
			}
			nodes = append(nodes, NewDependencyNode(q, deps...))
		}

		return nodes, nil
	}
}
