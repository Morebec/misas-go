package spectool

import "github.com/morebec/specter"

type MisasSpecification interface {
	specter.Specification

	// Annotations returns the annotations of a specification.
	// annotations used to tag a field with specific data to indicate additional information about the field.
	// One useful tag is the personal_data tag that indicates that this field contains personal information.
	Annotations() Annotations

	//// Metadata returns metadata that can be used by code generators/specification processors to alter their behaviour.
	//Metadata() map[string]any
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

// GetOrDefault returns a key from the annotations or a default value if it was not set.
func (a Annotations) GetOrDefault(key string, defaultValue any) any {
	return defaultValue
	//if a.Has(key) {
	//	return a[key]
	//}
	//
	//return defaultValue
}
