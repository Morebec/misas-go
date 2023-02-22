package specter

import "github.com/morebec/specter"

type MisasSpecification interface {
	specter.Specification

	// Annotations returns the annotations of a specification.
	// annotations used to tag a field with specific data to indicate additional information about the field.
	// One useful tag is the personal_data tag that indicates that this field contains personal information.
	Annotations() Annotations
}

// Annotations Represents a list of annotations.
type Annotations map[string]any

// HasKey indicates if the Annotations have a certain key.
func (a Annotations) HasKey(key string) bool {
	if _, ok := a[key]; ok {
		return true
	}

	return false
}

// GetOrDefault returns a key from the annotations or a default value if it was not set.
func (a Annotations) GetOrDefault(key string, defaultValue any) any {
	if a.HasKey(key) {
		return a[key]
	}

	return defaultValue
}
