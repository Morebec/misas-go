package spectool

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// SpecDeserializer is a service responsible for deserializing specs from SpecSource.
type SpecDeserializer interface {
	// Supports indicates if this deserializer supports a certain SpecSource.
	Supports(source SpecSource) bool

	// Deserialize a **supported** SpecSource to a Spec.
	// If the SpecSource is not supported, this method should return an error.
	// If one wants to ensure not to receive errors when calling this method, they should manually perform a call to Supports.
	Deserialize(source SpecSource) (Spec, error)
}

// TypeBasedDeserializer is an implementation of a SpecDeserializer that supports a given SpecType.
type TypeBasedDeserializer struct {
	supportedType SpecType
	deserialize   func(source SpecSource) (Spec, error)
}

func NewTypeBasedDeserializer(supportedType SpecType, deserialize func(source SpecSource) (Spec, error)) *TypeBasedDeserializer {
	return &TypeBasedDeserializer{supportedType: supportedType, deserialize: deserialize}
}

func (t TypeBasedDeserializer) Supports(source SpecSource) bool {
	yaml, err := parseYaml(source.Data)
	if err != nil {
		return false
	}
	specType := yaml["type"].(string)
	return SpecType(specType) == t.supportedType
}

func (t TypeBasedDeserializer) Deserialize(source SpecSource) (Spec, error) {
	if !t.Supports(source) {
		return Spec{}, errors.Errorf("could not deserialize spec of type %s at %s, as it is not supported.", source.SpecType, source.Location)
	}
	return t.deserialize(source)
}

// TypePropertiesDeserializer implementation of a SpecDeserializer that allows deserializing a spec and its properties.
type TypePropertiesDeserializer[T SpecProperties] struct {
	supportedType SpecType
}

func NewTypePropertiesDeserializer[T SpecProperties](supportedType SpecType) *TypePropertiesDeserializer[T] {
	return &TypePropertiesDeserializer[T]{supportedType: supportedType}
}

func (t TypePropertiesDeserializer[T]) Supports(source SpecSource) bool {
	parsedYaml, err := parseYaml(source.Data)
	if err != nil {
		return false
	}
	specType := parsedYaml["type"].(string)
	return SpecType(specType) == t.supportedType
}

func (t TypePropertiesDeserializer[T]) Deserialize(source SpecSource) (Spec, error) {
	if !t.Supports(source) {
		return Spec{}, errors.Errorf("could not deserialize spec of type %s at %s, as it is not supported.", source.SpecType, source.Location)
	}
	spec := Spec{}
	if err := yaml.Unmarshal(source.Data, &spec); err != nil {
		return Spec{}, errors.Wrapf(err, "failed deserializing command spec at %s", source.Location)
	}

	var props T
	if err := yaml.Unmarshal(source.Data, &props); err != nil {
		return Spec{}, errors.Wrapf(err, "failed deserializing command spec at %s", source.Location)
	}

	spec.Properties = props
	spec.Source = source

	return spec, nil
}

type CompositeDeserializer struct {
	deserializers []SpecDeserializer
}

func NewCompositeDeserializer(deserializers ...SpecDeserializer) *CompositeDeserializer {
	return &CompositeDeserializer{deserializers: deserializers}
}

func (a CompositeDeserializer) Supports(source SpecSource) bool {
	for _, d := range a.deserializers {
		if d.Supports(source) {
			return true
		}
	}

	return false
}

func (a CompositeDeserializer) Deserialize(source SpecSource) (Spec, error) {
	for _, d := range a.deserializers {
		if d.Supports(source) {
			s, err := d.Deserialize(source)
			if err != nil {
				return Spec{}, err
			}
			return s, nil
		}
	}

	return Spec{}, errors.Errorf("no deserializer supported for spec of type %s at %s", source.SpecType, source.Location)
}
