package spectool

import "github.com/pkg/errors"

const EnumType SpecType = "enum"

type EnumValue struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Value       any    `yaml:"value"`

	// Annotations are used to tag a field with specific data to indicate additional information about the field.
	// One useful tag is the personal_data tag that indicates that this field contains personal information.
	Annotations Annotations `yaml:"annotations"`
}
type EnumBaseType DataType

const StringEnum = EnumBaseType(String)
const IntEnum = EnumBaseType(Int)

type EnumSpecProperties struct {
	Values   map[string]EnumValue `yaml:"values"`
	BaseType EnumBaseType         `yaml:"baseType"`
}

func (c EnumSpecProperties) IsSpecProperties() {}

func EnumDeserializer() SpecDeserializer {
	inner := NewTypePropertiesDeserializer[EnumSpecProperties](EnumType)
	return NewTypeBasedDeserializer(EnumType, func(source SpecSource) (Spec, error) {
		spec, err := inner.Deserialize(source)
		if err != nil {
			return Spec{}, err
		}

		props := spec.Properties.(EnumSpecProperties)
		for vn, v := range props.Values {
			v.Name = vn
			props.Values[vn] = v
		}

		return spec, nil
	})
}

func EnumDependencyProvider() DependencyProvider {
	return func(systemSpec Spec, specs SpecGroup) ([]DependencyNode, error) {
		var nodes []DependencyNode
		return nodes, nil
	}
}

func EnumBaseTypeShouldBeSupportedEnumBaseType() Linter {
	return func(system Spec, specs SpecGroup) (LintingWarnings, LintingErrors) {
		var errs LintingErrors
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
//	return func(system Spec, specs SpecGroup) (LintingWarnings, LintingErrors) {
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
