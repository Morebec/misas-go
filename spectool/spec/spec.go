package spec

import (
	"github.com/morebec/misas-go/spectool/typesystem"
	"github.com/pkg/errors"
	"strings"
)

// Source represents a source file for a specification.
type Source struct {
	// Absolute path to the source.
	Location string

	// Content of the Source.
	Data []byte

	// Type of the spec of the source.
	SpecType Type
}

// Type Represents the "type" of a specification (e.g. Command, Struct, Query, Error)
type Type typesystem.Type

// TypeName represents the user defined type name of a Spec.
type TypeName typesystem.Type

// FollowsModuleConvention indicates if the type name follows the `module.type.other.other` convention.
func (t TypeName) FollowsModuleConvention() bool {
	return len(t.Parts()) >= 2
}

func (t TypeName) ModuleName() string {
	if !t.FollowsModuleConvention() {
		return ""
	}

	return t.Parts()[0]
}

// Parts returns the different parts in the type name. Parts are split by a "." character
func (t TypeName) Parts() []string {
	parts := strings.Split(string(t), ".")
	if len(parts) == 0 {
		return []string{
			string(t),
		}
	}
	return parts
}

// UndefinedSpecificationTypename constant used to test against undefined TypeName.
const UndefinedSpecificationTypename TypeName = ""

// UndefinedSpecificationType constant used to test against undefined Spec Type.
const UndefinedSpecificationType Type = ""

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

// Properties Represents the Type specific properties of a Spec.
type Properties interface {
	IsSpecProperties()
}

// Spec Represents a specification.
type Spec struct {
	// Type returns the Type of Spec.
	Type Type `yaml:"type"`

	// TypeName returns the TypeName of a Spec.
	TypeName TypeName `yaml:"typeName"`

	// Description returns the description of a Spec.
	Description string `yaml:"description"`

	// Annotations return the annotations of a Specifications.
	Annotations Annotations `yaml:"annotations"`

	// Version Returns the version of the spec.
	Version string `yaml:"version"`

	// Source returns the SpecificationSource of a Spec.
	Source Source

	Properties Properties
}

// Group Represents a list of Spec.
type Group []Spec

// Merge Allows merging a group with another one.
func (g Group) Merge(group Group) Group {
	merged := g
	typeNameIndex := map[TypeName]any{}
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
func (g Group) Select(p func(s Spec) bool) Group {
	r := Group{}
	for _, s := range g {
		if p(s) {
			r = append(r, s)
		}
	}

	return r
}

func (g Group) SelectType(t Type) Group {
	return g.Select(func(s Spec) bool {
		return s.Type == t
	})
}

func (g Group) SelectTypeName(t TypeName) Spec {
	for _, s := range g {
		if s.TypeName == t {
			return s
		}
	}

	return Spec{}
}

func (g Group) Exclude(p func(s Spec) bool) Group {
	r := Group{}
	for _, s := range g {
		if !p(s) {
			r = append(r, s)
		}
	}

	return r
}

func (g Group) ExcludeType(t Type) Group {
	return g.Exclude(func(s Spec) bool {
		return s.Type == t
	})
}

// MapSpecGroup performs a map operation on a Group
func MapSpecGroup[T any](g Group, p func(s Spec) T) []T {
	var mapped []T
	for _, s := range g {
		mapped = append(mapped, p(s))
	}

	return mapped
}

func UnexpectedSpecTypeError(actual Type, expected Type) error {
	return errors.Errorf("expected spec of type %s, got %s", expected, actual)
}
