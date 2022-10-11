package builtin

import (
	"fmt"
	"github.com/morebec/misas-go/spectool/processing"
	"github.com/morebec/misas-go/spectool/spec"
	"github.com/morebec/misas-go/spectool/typesystem"
	"github.com/pkg/errors"
)

const EventType spec.Type = "event"

type EventField struct {
	Name        string              `yaml:"name"`
	Description string              `yaml:"description"`
	Type        typesystem.DataType `yaml:"type"`
	Nullable    bool                `yaml:"nullable"`
	Deprecation string              `yaml:"deprecation"`
	Example     string              `yaml:"example"`

	// Annotations are used to tag a field with specific data to indicate additional information about the field.
	// One useful tag is the personal_data tag that indicates that this field contains personal information.
	Annotations spec.Annotations `yaml:"annotations"`
}
type EventSpecProperties struct {
	Fields map[string]EventField `yaml:"fields"`
}

func (c EventSpecProperties) IsSpecProperties() {}

func EventDeserializer() processing.SpecDeserializer {
	inner := processing.NewTypePropertiesDeserializer[EventSpecProperties](EventType)
	return processing.NewTypeBasedDeserializer(EventType, func(source spec.Source) (spec.Spec, error) {
		s, err := inner.Deserialize(source)
		if err != nil {
			return spec.Spec{}, err
		}

		props := s.Properties.(EventSpecProperties)
		for fn, f := range props.Fields {
			f.Name = fn
			props.Fields[fn] = f
		}

		return s, nil
	})
}

func EventDependencyProvider() processing.DependencyProvider {
	return func(systemSpec spec.Spec, specs spec.Group) ([]processing.DependencyNode, error) {
		events := specs.SelectType(EventType)

		var nodes []processing.DependencyNode

		for _, e := range events {
			props := e.Properties.(EventSpecProperties)
			var deps []spec.TypeName
			for _, f := range props.Fields {
				userDefined := f.Type.ExtractUserDefined()
				if userDefined != "" {
					deps = append(deps, spec.TypeName(userDefined))
				}
			}
			nodes = append(nodes, processing.NewDependencyNode(e, deps...))
		}

		return nodes, nil
	}
}

func EventsMustHaveDateTimeField() processing.Linter {
	return func(system spec.Spec, specs spec.Group) (processing.LintingWarnings, processing.LintingErrors) {
		events := specs.SelectType(EventType)
		var lintingErrors processing.LintingErrors
		for _, e := range events {
			props := e.Properties.(EventSpecProperties)
			fieldFound := false
			for _, f := range props.Fields {
				if f.Type == typesystem.DateTime {
					fieldFound = true
					break
				}
			}
			if !fieldFound {
				lintingErrors = append(lintingErrors, errors.Errorf("event %s does not have a date time field at %s", e.TypeName, e.Source.Location))
			}
		}

		return nil, lintingErrors
	}
}

func EventFieldsShouldHaveDescriptionLinter() processing.Linter {
	return func(system spec.Spec, specs spec.Group) (processing.LintingWarnings, processing.LintingErrors) {
		var warning processing.LintingWarnings
		for _, s := range specs.SelectType(EventType) {
			props := s.Properties.(EventSpecProperties)
			for _, f := range props.Fields {
				if f.Description == "" {
					warning = append(warning, fmt.Sprintf("field %s of event %s does not have a description", f.Name, s.TypeName))
				}
			}
		}
		return warning, nil
	}
}
