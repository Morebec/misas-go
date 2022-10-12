package builtin

import (
	"fmt"
	"github.com/morebec/misas-go/spectool/processing"
	"github.com/morebec/misas-go/spectool/spec"
	"github.com/morebec/misas-go/spectool/typesystem"
	"github.com/pkg/errors"
)

const QueryType spec.Type = "query"

type QueryField struct {
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
type QuerySpecProperties struct {
	Fields map[string]QueryField `yaml:"fields"`
}

func (c QuerySpecProperties) IsSpecProperties() {}

func QueryDeserializer() processing.SpecDeserializer {
	inner := processing.NewTypePropertiesDeserializer[QuerySpecProperties](QueryType)
	return processing.NewTypeBasedDeserializer(QueryType, func(source spec.Source) (spec.Spec, error) {
		s, err := inner.Deserialize(source)
		if err != nil {
			return spec.Spec{}, err
		}

		props := s.Properties.(QuerySpecProperties)
		for fn, f := range props.Fields {
			f.Name = fn
			props.Fields[fn] = f
		}

		return s, nil
	})
}

func QueryDependencyProvider() processing.DependencyProvider {
	return func(systemSpec spec.Spec, specs spec.Group) ([]processing.DependencyNode, error) {
		queries := specs.SelectType(QueryType)

		var nodes []processing.DependencyNode

		for _, q := range queries {
			props := q.Properties.(QuerySpecProperties)
			var deps []spec.TypeName
			for _, f := range props.Fields {
				userDefined := f.Type.ExtractUserDefined()
				if userDefined != "" {
					deps = append(deps, spec.TypeName(userDefined))
				}
			}
			nodes = append(nodes, processing.NewDependencyNode(q, deps...))
		}

		return nodes, nil
	}
}

func QueryFieldsMustHaveTypeLinter() processing.Linter {
	return func(system spec.Spec, specs spec.Group) (processing.LintingWarnings, processing.LintingErrors) {
		var errs processing.LintingErrors
		for _, s := range specs.SelectType(QueryType) {
			props := s.Properties.(QuerySpecProperties)
			for _, f := range props.Fields {
				if f.Type == "" {
					errs = append(errs, errors.Errorf("field %s of query %s does not have a type", f.Name, s.TypeName))
				}
			}
		}
		return nil, errs
	}
}

func QueryFieldsShouldHaveDescriptionLinter() processing.Linter {
	return func(system spec.Spec, specs spec.Group) (processing.LintingWarnings, processing.LintingErrors) {
		var warning processing.LintingWarnings
		for _, s := range specs.SelectType(QueryType) {
			props := s.Properties.(QuerySpecProperties)
			for _, f := range props.Fields {
				if f.Description == "" {
					warning = append(warning, fmt.Sprintf("field %s of query %s does not have a description", f.Name, s.TypeName))
				}
			}
		}
		return warning, nil
	}
}
