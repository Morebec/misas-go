package spectool

import (
	"fmt"
	"github.com/pkg/errors"
)

const CommandType SpecType = "command"

type CommandField struct {
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
type CommandSpecProperties struct {
	Fields map[string]CommandField `yaml:"fields"`
}

func (c CommandSpecProperties) IsSpecProperties() {}

func CommandDeserializer() SpecDeserializer {
	inner := NewTypePropertiesDeserializer[CommandSpecProperties](CommandType)
	return NewTypeBasedDeserializer(CommandType, func(source SpecSource) (Spec, error) {
		spec, err := inner.Deserialize(source)
		if err != nil {
			return Spec{}, err
		}

		props := spec.Properties.(CommandSpecProperties)
		for fn, f := range props.Fields {
			f.Name = fn
			props.Fields[fn] = f
		}

		return spec, nil
	})
}

func CommandDependencyProvider() DependencyProvider {
	return func(systemSpec Spec, specs SpecGroup) ([]DependencyNode, error) {
		commands := specs.SelectType(CommandType)

		var nodes []DependencyNode

		for _, cmd := range commands {
			props := cmd.Properties.(CommandSpecProperties)
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

func CommandFieldsMustHaveNameLinter() Linter {
	return func(system Spec, specs SpecGroup) (LintingWarnings, LintingErrors) {
		var errs LintingErrors
		for _, s := range specs.SelectType(CommandType) {
			props := s.Properties.(CommandSpecProperties)
			for i, f := range props.Fields {
				if f.Description == "" {
					errs = append(errs, errors.Errorf("field [%s] of command %s does not have a name", i, s.TypeName))
				}
			}
		}
		return nil, errs
	}
}

func CommandFieldsShouldHaveDescriptionLinter() Linter {
	return func(system Spec, specs SpecGroup) (LintingWarnings, LintingErrors) {
		var warning LintingWarnings
		for _, s := range specs.SelectType(CommandType) {
			props := s.Properties.(CommandSpecProperties)
			for _, f := range props.Fields {
				if f.Description == "" {
					warning = append(warning, fmt.Sprintf("field %s of command %s does not have a description", f.Name, s.TypeName))
				}
			}
		}
		return warning, nil
	}
}
