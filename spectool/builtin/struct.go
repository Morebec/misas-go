package builtin

import (
	"fmt"
	"github.com/morebec/misas-go/spectool/processing"
	"github.com/morebec/misas-go/spectool/spec"
	"github.com/morebec/misas-go/spectool/typesystem"
)

const StructType spec.Type = "struct"

type StructField struct {
	Name        string              `yaml:"name"`
	Description string              `yaml:"description"`
	Type        typesystem.DataType `yaml:"type"`
	Nullable    bool                `yaml:"nullable"`
	Deprecation string              `yaml:"deprecation"`
	Example     string              `yaml:"example"`
	Default     string              `yaml:"default"`
	Required    bool                `yaml:"required"`

	// Annotations are used to tag a field with specific data to indicate additional information about the field.
	// One useful tag is the personal_data tag that indicates that this field contains personal information.
	Annotations spec.Annotations `yaml:"annotations"`
}
type StructSpecProperties struct {
	Fields map[string]StructField `yaml:"fields"`
}

func (c StructSpecProperties) IsSpecProperties() {}

func StructDeserializer() processing.SpecDeserializer {
	inner := processing.NewTypePropertiesDeserializer[StructSpecProperties](StructType)
	return processing.NewTypeBasedDeserializer(StructType, func(source spec.Source) (spec.Spec, error) {
		s, err := inner.Deserialize(source)
		if err != nil {
			return spec.Spec{}, err
		}

		props := s.Properties.(StructSpecProperties)
		for fn, f := range props.Fields {
			f.Name = fn
			props.Fields[fn] = f
		}

		return s, nil
	})
}

func StructDependencyProvider() processing.DependencyProvider {
	return func(systemSpec spec.Spec, specs spec.Group) ([]processing.DependencyNode, error) {
		structs := specs.SelectType(StructType)

		var nodes []processing.DependencyNode

		for _, cmd := range structs {
			props := cmd.Properties.(StructSpecProperties)
			var deps []spec.TypeName
			for _, f := range props.Fields {
				userDefined := f.Type.ExtractUserDefined()
				if userDefined != "" {
					deps = append(deps, spec.TypeName(userDefined))
				}
			}
			nodes = append(nodes, processing.NewDependencyNode(cmd, deps...))
		}

		return nodes, nil
	}
}

func StructFieldsShouldHaveDescriptionLinter() processing.Linter {
	return func(system spec.Spec, specs spec.Group) (processing.LintingWarnings, processing.LintingErrors) {
		var warning processing.LintingWarnings
		for _, s := range specs.SelectType(StructType) {
			props := s.Properties.(StructSpecProperties)
			for _, f := range props.Fields {
				if f.Description == "" {
					warning = append(warning, fmt.Sprintf("field %s of struct %s does not have a description", f.Name, s.TypeName))
				}
			}
		}
		return warning, nil
	}
}
