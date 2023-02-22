package misas

import (
	"github.com/morebec/misas-go/specter/v1/spec"
)

type CommandField struct {
	Name        string
	Description string
	Type        DataType
	Nullable    bool
	Deprecation string
	Example     string
	Default     string
	Required    bool

	// Annotations are used to tag a field with specific data to indicate additional information about the field.
	// One useful tag is the personal_data tag that indicates that this field contains personal information.
	// Annotations spec.Meta
}

type Command struct {
	Fields []CommandField
	Spec   spec.Spec
}

func (c Command) Dependencies() []spec.Name {
	var deps []spec.Name
	for _, f := range c.Fields {
		if f.Type.IsUserDefined() {
			deps = append(deps, spec.Name(f.Type))
		}
	}
	return deps
}

func (c Command) Name() spec.Name {
	return c.Spec.Name
}

type CommandFactory struct {
}

func (c CommandFactory) Make(s spec.Spec) (spec.Schema, error) {
	if s.Type != "command" {
		// TODO return error
	}

	//var fields []CommandField
	//for _, a := range s.Attributes {
	//	data, ok := a.Value.(spec.ObjectValue)
	//	if !ok {
	//		continue
	//	}
	//	if data.Type != "field" {
	//		continue
	//	}
	//}

	return nil, nil
}
