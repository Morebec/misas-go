package spectool

import (
	"fmt"
	"github.com/pkg/errors"
)

const EventType SpecType = "event"

type EventField struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Type        DataType `yaml:"type"`
	Nullable    bool     `yaml:"nullable"`
	Deprecation string   `yaml:"deprecation"`
	Example     string   `yaml:"example"`

	// Annotations are used to tag a field with specific data to indicate additional information about the field.
	// One useful tag is the personal_data tag that indicates that this field contains personal information.
	Annotations Annotations `yaml:"annotations"`
}
type EventSpecProperties struct {
	Fields map[string]EventField `yaml:"fields"`
}

func (c EventSpecProperties) IsSpecProperties() {}

func EventDeserializer() SpecDeserializer {
	inner := NewTypePropertiesDeserializer[EventSpecProperties](EventType)
	return NewTypeBasedDeserializer(EventType, func(source SpecSource) (Spec, error) {
		spec, err := inner.Deserialize(source)
		if err != nil {
			return Spec{}, err
		}

		props := spec.Properties.(EventSpecProperties)
		for fn, f := range props.Fields {
			f.Name = fn
			props.Fields[fn] = f
		}

		return spec, nil
	})
}

func EventDependencyProvider() DependencyProvider {
	return func(systemSpec Spec, specs SpecGroup) ([]DependencyNode, error) {
		events := specs.SelectType(EventType)

		var nodes []DependencyNode

		for _, e := range events {
			props := e.Properties.(EventSpecProperties)
			var deps []SpecTypeName
			for _, f := range props.Fields {
				userDefined := f.Type.ExtractUserDefined()
				if userDefined != "" {
					deps = append(deps, SpecTypeName(userDefined))
				}
			}
			nodes = append(nodes, NewDependencyNode(e, deps...))
		}

		return nodes, nil
	}
}

func EventsMustHaveDateTimeField() Linter {
	return func(system Spec, specs SpecGroup) (LintingWarnings, LintingErrors) {
		events := specs.SelectType(EventType)
		var lintingErrors LintingErrors
		for _, e := range events {
			props := e.Properties.(EventSpecProperties)
			fieldFound := false
			for _, f := range props.Fields {
				if f.Type == DateTime {
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

func EventFieldsMustHaveNameLinter() Linter {
	return func(system Spec, specs SpecGroup) (LintingWarnings, LintingErrors) {
		var errs LintingErrors
		for _, s := range specs.SelectType(EventType) {
			props := s.Properties.(EventSpecProperties)
			for i, f := range props.Fields {
				if f.Description == "" {
					errs = append(errs, errors.Errorf("field [%s] of event %s does not have a name", i, f.Description))
				}
			}
		}
		return nil, errs
	}
}

func EventFieldsShouldHaveDescriptionLinter() Linter {
	return func(system Spec, specs SpecGroup) (LintingWarnings, LintingErrors) {
		var warning LintingWarnings
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
