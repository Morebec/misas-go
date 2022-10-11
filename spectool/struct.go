package spectool

import (
	"fmt"
	"github.com/pkg/errors"
)

const StructType SpecType = "struct"

type StructField struct {
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
type StructSpecProperties struct {
	Fields map[string]StructField `yaml:"fields"`
}

func (c StructSpecProperties) IsSpecProperties() {}

func StructDeserializer() SpecDeserializer {
	inner := NewTypePropertiesDeserializer[StructSpecProperties](StructType)
	return NewTypeBasedDeserializer(StructType, func(source SpecSource) (Spec, error) {
		spec, err := inner.Deserialize(source)
		if err != nil {
			return Spec{}, err
		}

		props := spec.Properties.(StructSpecProperties)
		for fn, f := range props.Fields {
			f.Name = fn
			props.Fields[fn] = f
		}

		return spec, nil
	})
}

func StructDependencyProvider() DependencyProvider {
	return func(systemSpec Spec, specs SpecGroup) ([]DependencyNode, error) {
		structs := specs.SelectType(StructType)

		var nodes []DependencyNode

		for _, cmd := range structs {
			props := cmd.Properties.(StructSpecProperties)
			var deps []SpecTypeName
			for _, f := range props.Fields {
				userDefined := f.Type.ExtractUserDefined()
				if userDefined != "" {
					deps = append(deps, SpecTypeName(userDefined))
				}
			}
			nodes = append(nodes, NewDependencyNode(cmd, deps...))
		}

		return nodes, nil
	}
}

func StructFieldsMustHaveNameLinter() Linter {
	return func(system Spec, specs SpecGroup) (LintingWarnings, LintingErrors) {
		var errs LintingErrors
		for _, s := range specs.SelectType(StructType) {
			props := s.Properties.(StructSpecProperties)
			for i, f := range props.Fields {
				if f.Description == "" {
					errs = append(errs, errors.Errorf("field [%s] of struct %s does not have a name", i, f.Description))
				}
			}
		}
		return nil, errs
	}
}

func StructFieldsShouldHaveDescriptionLinter() Linter {
	return func(system Spec, specs SpecGroup) (LintingWarnings, LintingErrors) {
		var warning LintingWarnings
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
