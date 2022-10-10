// Copyright 2022 Morébec
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prediction

import (
	"encoding/json"
	"github.com/pkg/errors"
	"reflect"
)

// Converter is a service responsible for converting Prediction to Descriptor and back.
// Internally it relies on mapping the empty value of a Prediction to its TypeName so that it can read the TypeName
// of a given Descriptor to have the right in memory representation (struct) of the Prediction.
type Converter struct {
	events map[TypeName]reflect.Type
}

func NewConverter() *Converter {
	return &Converter{map[TypeName]reflect.Type{}}
}

// FromDescriptor loads a Prediction from a Descriptor
func (e *Converter) FromDescriptor(descriptor Descriptor) (Prediction, error) {
	evt, err := e.find(descriptor.TypeName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed loading prediction \"%s/%s\"", descriptor.TypeName, descriptor.ID)
	}

	marshal, err := json.Marshal(descriptor.Payload)
	if err != nil {
		return nil, errors.Wrapf(err, "failed loading prediction \"%s/%s\"", descriptor.TypeName, descriptor.ID)
	}

	if err := json.Unmarshal(marshal, evt); err != nil {
		return nil, errors.Wrapf(err, "failed loading prediction \"%s/%s\"", descriptor.TypeName, descriptor.ID)
	}

	// Convert *Prediction to Prediction
	return reflect.ValueOf(evt).Elem().Interface().(Prediction), nil
}

// RegisterPrediction registers an event and its type with this loader.
func (e *Converter) RegisterPrediction(c Prediction) *Converter {
	typ := reflect.TypeOf(c)
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	e.events[c.TypeName()] = typ
	return e
}

// find a pointer to a struct of the event's type to be used for loading.
func (e *Converter) find(tn TypeName) (Prediction, error) {
	evt, found := e.events[tn]
	if !found {
		return nil, errors.Errorf("prediction event registered for type name \"%s\"", tn)
	}

	// Create a new instance
	ret := reflect.ValueOf(reflect.New(evt).Interface()).Elem().Interface().(Prediction)

	// Create a pointer since most decoders require pointer values.
	evtPtr := reflect.New(reflect.TypeOf(ret)).Interface().(Prediction)
	return evtPtr, nil
}
