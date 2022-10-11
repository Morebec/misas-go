package processing

import (
	"github.com/morebec/misas-go/spectool/spec"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// SpecDeserializer is a service responsible for deserializing specs from spec.Source.
type SpecDeserializer interface {
	// Supports indicates if this deserializer supports a certain spec.Source.
	Supports(source spec.Source) bool

	// Deserialize a **supported** spec.Source to a Spec.
	// If the spec.Source is not supported, this method should return an error.
	// If one wants to ensure not to receive errors when calling this method, they should manually perform a call to Supports.
	Deserialize(source spec.Source) (spec.Spec, error)
}

// TypeBasedDeserializer is an implementation of a SpecDeserializer that supports a given Type.
type TypeBasedDeserializer struct {
	supportedType spec.Type
	deserialize   func(source spec.Source) (spec.Spec, error)
}

func NewTypeBasedDeserializer(supportedType spec.Type, deserialize func(source spec.Source) (spec.Spec, error)) *TypeBasedDeserializer {
	return &TypeBasedDeserializer{supportedType: supportedType, deserialize: deserialize}
}

func (t TypeBasedDeserializer) Supports(source spec.Source) bool {
	data, err := parseYaml(source.Data)
	if err != nil {
		return false
	}
	specType := data["type"].(string)
	return spec.Type(specType) == t.supportedType
}

func (t TypeBasedDeserializer) Deserialize(source spec.Source) (spec.Spec, error) {
	if !t.Supports(source) {
		return spec.Spec{}, errors.Errorf("could not deserialize spec of type %s at %s, as it is not supported.", source.SpecType, source.Location)
	}
	return t.deserialize(source)
}

// TypePropertiesDeserializer implementation of a SpecDeserializer that allows deserializing a spec and its properties.
type TypePropertiesDeserializer[T spec.Properties] struct {
	supportedType spec.Type
}

func NewTypePropertiesDeserializer[T spec.Properties](supportedType spec.Type) *TypePropertiesDeserializer[T] {
	return &TypePropertiesDeserializer[T]{supportedType: supportedType}
}

func (t TypePropertiesDeserializer[T]) Supports(source spec.Source) bool {
	parsedYaml, err := parseYaml(source.Data)
	if err != nil {
		return false
	}
	specType := parsedYaml["type"].(string)
	return spec.Type(specType) == t.supportedType
}

func (t TypePropertiesDeserializer[T]) Deserialize(source spec.Source) (spec.Spec, error) {
	if !t.Supports(source) {
		return spec.Spec{}, errors.Errorf("could not deserialize spec of type %s at %s, as it is not supported.", source.SpecType, source.Location)
	}
	s := spec.Spec{}
	if err := yaml.Unmarshal(source.Data, &s); err != nil {
		return spec.Spec{}, errors.Wrapf(err, "failed deserializing command spec at %s", source.Location)
	}

	var props T
	if err := yaml.Unmarshal(source.Data, &props); err != nil {
		return spec.Spec{}, errors.Wrapf(err, "failed deserializing command spec at %s", source.Location)
	}

	s.Properties = props
	s.Source = source

	return s, nil
}

type CompositeDeserializer struct {
	deserializers []SpecDeserializer
}

func NewCompositeDeserializer(deserializers ...SpecDeserializer) *CompositeDeserializer {
	return &CompositeDeserializer{deserializers: deserializers}
}

func (a CompositeDeserializer) Supports(source spec.Source) bool {
	for _, d := range a.deserializers {
		if d.Supports(source) {
			return true
		}
	}

	return false
}

func (a CompositeDeserializer) Deserialize(source spec.Source) (spec.Spec, error) {
	for _, d := range a.deserializers {
		if d.Supports(source) {
			s, err := d.Deserialize(source)
			if err != nil {
				return spec.Spec{}, err
			}
			return s, nil
		}
	}

	return spec.Spec{}, errors.Errorf("no deserializer supported for spec of type %s at %s", source.SpecType, source.Location)
}
