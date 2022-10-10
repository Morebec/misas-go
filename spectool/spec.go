package spectool

import (
	"github.com/pkg/errors"
	"strings"
)

// SpecType Represents the "type" of a specification (e.g. Command, Struct, Query, Error)
type SpecType Type

// SpecTypeName represents the user defined type name of a Spec.
type SpecTypeName Type

// FollowsModuleConvention indicates if the type name follows the `module.type.other.other` convention.
func (t SpecTypeName) FollowsModuleConvention() bool {
	return len(t.Parts()) >= 2
}

func (t SpecTypeName) ModuleName() string {
	if !t.FollowsModuleConvention() {
		return ""
	}

	return t.Parts()[0]
}

// Parts returns the different parts in the type name. Parts are split by a "." character
func (t SpecTypeName) Parts() []string {
	parts := strings.Split(string(t), ".")
	if len(parts) == 0 {
		return []string{
			string(t),
		}
	}
	return parts
}

// UndefinedSpecificationTypename constant used to test against undefined SpecTypeName.
const UndefinedSpecificationTypename SpecTypeName = ""

// UndefinedSpecificationType constant used to test against undefined Spec SpecType.
const UndefinedSpecificationType SpecType = ""

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

// SpecProperties Represents the SpecType specific properties of a Spec.
type SpecProperties interface {
	IsSpecProperties()
}

// Spec Represents a specification.
type Spec struct {
	// Type returns the SpecType of a Spec.
	Type SpecType `yaml:"type"`

	// TypeName returns the SpecTypeName of a Spec.
	TypeName SpecTypeName `yaml:"typeName"`

	// Description returns the description of a Spec.
	Description string `yaml:"description"`

	// Annotations return the annotations of a Specifications.
	Annotations Annotations `yaml:"annotations"`

	// Version Returns the version of the spec.
	Version string `yaml:"version"`

	// Source returns the SpecificationSource of a Spec.
	Source SpecSource

	Properties SpecProperties
}

// SpecGroup Represents a list of Spec.
type SpecGroup []Spec

// Merge Allows merging a group with another one.
func (g SpecGroup) Merge(group SpecGroup) SpecGroup {
	merged := g
	typeNameIndex := map[SpecTypeName]any{}
	for _, s := range g {
		typeNameIndex[s.TypeName] = nil
	}
	for _, s := range group {
		if _, found := typeNameIndex[s.TypeName]; found {
			continue
		}
		typeNameIndex[s.TypeName] = nil
		merged = append(merged, s)
	}
	return merged
}

// Select allows filtering the group for certain specifications.
func (g SpecGroup) Select(p func(s Spec) bool) SpecGroup {
	r := SpecGroup{}
	for _, s := range g {
		if p(s) {
			r = append(r, s)
		}
	}

	return r
}

func (g SpecGroup) SelectType(t SpecType) SpecGroup {
	return g.Select(func(s Spec) bool {
		return s.Type == t
	})
}

func (g SpecGroup) SelectTypeName(t SpecTypeName) Spec {
	for _, s := range g {
		if s.TypeName == t {
			return s
		}
	}

	return Spec{}
}

func (g SpecGroup) Exclude(p func(s Spec) bool) SpecGroup {
	r := SpecGroup{}
	for _, s := range g {
		if !p(s) {
			r = append(r, s)
		}
	}

	return r
}

func (g SpecGroup) ExcludeType(t SpecType) SpecGroup {
	return g.Exclude(func(s Spec) bool {
		return s.Type == t
	})
}

// MapSpecGroup performs a map operation on a SpecGroup
func MapSpecGroup[T any](g SpecGroup, p func(s Spec) T) []T {
	var mapped []T
	for _, s := range g {
		mapped = append(mapped, p(s))
	}

	return mapped
}

func UnexpectedSpecTypeError(actual SpecType, expected SpecType) error {
	return errors.Errorf("expected spec of type %s, got %s", expected, actual)
}
