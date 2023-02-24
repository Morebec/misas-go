package spectool

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/morebec/specter"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type MetadataEntry struct {
	Key   string         `hcl:"key,label"`
	Value *hcl.Attribute `hcl:"value"`
}

type Metadata []MetadataEntry

func (m Metadata) GetOrDefault(key string, defaultValue any) cty.Value {
	if m != nil {

		for _, e := range m {
			if e.Key == key {
				if value, err := e.Value.Expr.Value(&hcl.EvalContext{}); err != nil {
					panic(err)
				} else {
					return value
				}
			}
		}
	}

	var ty cty.Type
	switch defaultValue.(type) {
	case string:
		ty = cty.String
	case bool:
		ty = cty.Bool
	case int, int32, int64:
		ty = cty.Number
	case float32, float64:
		ty = cty.Number
	default:
		ty = cty.DynamicPseudoType
	}
	if defVal, err := gocty.ToCtyValue(defaultValue, ty); err != nil {
		panic(err)
	} else {
		return defVal
	}
}

func (m Metadata) HasKey(key string) bool {
	if m != nil {
		for _, e := range m {
			if e.Key == key {
				return true
			}
		}
	}
	return false
}

type MisasSpecification interface {
	specter.Specification

	// Annotations returns the annotations of a specification.
	// annotations used to tag a field with specific data to indicate additional information about the field.
	// One useful tag is the personal_data tag that indicates that this field contains personal information.
	Annotations() Annotations

	// Metadata returns metadata that can be used by code generators/specification processors to alter their behaviour.
	Metadata() Metadata
}

// Annotations Represents a list of annotations.
type Annotations []string

// Has indicates if the Annotations have a certain key.
func (a Annotations) Has(value string) bool {
	for _, v := range a {
		if v == value {
			return true
		}
	}

	return false
}
