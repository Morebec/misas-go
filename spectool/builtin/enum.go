package builtin

import (
	"github.com/morebec/misas-go/spectool/processing"
	"github.com/morebec/misas-go/spectool/spec"
	"github.com/morebec/misas-go/spectool/typesystem"
	"github.com/pkg/errors"
)

const EnumType spec.Type = "enum"

type EnumValue struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Value       any    `yaml:"value"`

	// Annotations are used to tag a field with specific data to indicate additional information about the field.
	// One useful tag is the personal_data tag that indicates that this field contains personal information.
	Annotations spec.Annotations `yaml:"annotations"`
}
type EnumBaseType typesystem.DataType

const StringEnum = EnumBaseType(typesystem.String)
const IntEnum = EnumBaseType(typesystem.Int)

type EnumSpecProperties struct {
	Values   map[string]EnumValue `yaml:"values"`
	BaseType EnumBaseType         `yaml:"baseType"`
}

func (c EnumSpecProperties) IsSpecProperties() {}

func EnumDeserializer() processing.SpecDeserializer {
	inner := processing.NewTypePropertiesDeserializer[EnumSpecProperties](EnumType)
	return processing.NewTypeBasedDeserializer(EnumType, func(source spec.Source) (spec.Spec, error) {
		s, err := inner.Deserialize(source)
		if err != nil {
			return spec.Spec{}, err
		}

		props := s.Properties.(EnumSpecProperties)
		for vn, v := range props.Values {
			v.Name = vn
			props.Values[vn] = v
		}

		return s, nil
	})
}

func EnumDependencyProvider() processing.DependencyProvider {
	return func(systemSpec spec.Spec, specs spec.Group) ([]processing.DependencyNode, error) {
		var nodes []processing.DependencyNode
		return nodes, nil
	}
}

func EnumBaseTypeShouldBeSupportedEnumBaseType() processing.Linter {
	return func(system spec.Spec, specs spec.Group) (processing.LintingWarnings, processing.LintingErrors) {
		var errs processing.LintingErrors
		enums := specs.SelectType(EnumType)
		for _, e := range enums {
			props := e.Properties.(EnumSpecProperties)
			if props.BaseType != StringEnum && props.BaseType != IntEnum {
				errs = append(errs, errors.Errorf(
					"Enum %s does not have a valid base type expected, %s, %s, got %s at %s",
					e.TypeName,
					StringEnum,
					IntEnum,
					props.BaseType,
					e.Source.Location,
				))
			}
		}
		return nil, errs
	}
}

//func EnumValueNamesShouldBeUnique() Linter {
//	return func(system Spec, specs Group) (LintingWarnings, LintingErrors) {
//		var errors LintingErrors
//		enums := specs.SelectType(EnumType)
//		var names map[string]struct{}
//		for _, e := range enums {
//			for _, v := range e.Properties.(EnumSpecProperties).Values {
//				if _, found := names[v.Name]; found {
//					errors = append(errors, "Enum %s has duplicate value names for")
//				}
//			}
//		}
//
//	}
//}
