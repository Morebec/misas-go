package builtin

import (
	"fmt"
	"github.com/morebec/misas-go/spectool/processing"
	"github.com/morebec/misas-go/spectool/spec"
	"github.com/morebec/misas-go/spectool/typesystem"
	"github.com/pkg/errors"
)

const CommandType spec.Type = "command"

type CommandField struct {
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
type CommandSpecProperties struct {
	Fields map[string]CommandField `yaml:"fields"`
}

func (c CommandSpecProperties) IsSpecProperties() {}

func CommandDeserializer() processing.SpecDeserializer {
	inner := processing.NewTypePropertiesDeserializer[CommandSpecProperties](CommandType)
	return processing.NewTypeBasedDeserializer(CommandType, func(source spec.Source) (spec.Spec, error) {
		s, err := inner.Deserialize(source)
		if err != nil {
			return spec.Spec{}, err
		}

		props := s.Properties.(CommandSpecProperties)
		for fn, f := range props.Fields {
			f.Name = fn
			props.Fields[fn] = f
		}

		return s, nil
	})
}

func CommandDependencyProvider() processing.DependencyProvider {
	return func(systemSpec spec.Spec, specs spec.Group) ([]processing.DependencyNode, error) {
		commands := specs.SelectType(CommandType)

		var nodes []processing.DependencyNode

		for _, cmd := range commands {
			props := cmd.Properties.(CommandSpecProperties)
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

func CommandFieldsMustHaveTypeLinter() processing.Linter {
	return func(system spec.Spec, specs spec.Group) (processing.LintingWarnings, processing.LintingErrors) {
		var errs processing.LintingErrors
		for _, s := range specs.SelectType(CommandType) {
			props := s.Properties.(CommandSpecProperties)
			for _, f := range props.Fields {
				if f.Type == "" {
					errs = append(errs, errors.Errorf("field %s of command %s does not have a type", f.Name, s.TypeName))
				}
			}
		}
		return nil, errs
	}
}

func CommandFieldsShouldHaveDescriptionLinter() processing.Linter {
	return func(system spec.Spec, specs spec.Group) (processing.LintingWarnings, processing.LintingErrors) {
		var warning processing.LintingWarnings
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
