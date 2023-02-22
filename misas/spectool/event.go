package spectool

import (
	"fmt"
	"github.com/morebec/specter"
)

type EventField struct {
	Name        string   `hcl:"name,label"`
	Description string   `hcl:"description"`
	Type        DataType `hcl:"type"`
	Nullable    bool     `hcl:"nullable,optional"`
	Deprecation string   `hcl:"deprecation,optional"`
	Example     string   `hcl:"example,optional"`

	// Annotations are used to tag a field with specific data to indicate additional information about the field.
	// One useful tag is the personal_data tag that indicates that this field contains personal information.
	Annotations []string `hcl:"annotations,optional"`
}

type Event struct {
	Nam    string       `hcl:"name,label"`
	Desc   string       `hcl:"description"`
	Fields []EventField `hcl:"field,block"`
	Src    specter.Source
	Annots Annotations `hcl:"annotations,optional"`
}

func (e *Event) Annotations() Annotations {
	return e.Annots
}

func (e *Event) Name() specter.SpecificationName {
	return specter.SpecificationName(e.Nam)
}

func (e *Event) Type() specter.SpecificationType {
	return "event"
}

func (e *Event) Description() string {
	return e.Desc
}

func (e *Event) Source() specter.Source {
	return e.Src
}

func (e *Event) SetSource(s specter.Source) {
	e.Src = s
}

func (e *Event) Dependencies() []specter.SpecificationName {
	var deps []specter.SpecificationName
	for _, f := range e.Fields {
		if f.Type.IsUserDefined() {
			deps = append(deps, specter.SpecificationName(f.Type))
		}
	}
	return deps
}

func EventsMustHaveDateTimeField() specter.SpecificationLinterFunc {
	return func(specs specter.SpecificationGroup) specter.LinterResultSet {
		events := specs.SelectType(specter.SpecificationType("event"))
		var result specter.LinterResultSet
		for _, e := range events {
			evt := e.(*Event)
			fieldFound := false
			for _, f := range evt.Fields {
				if f.Type == DateTime {
					fieldFound = true
					break
				}
			}
			if !fieldFound {
				result = append(result, specter.LinterResult{
					Severity: specter.ErrorSeverity,
					Message:  fmt.Sprintf("event \"%s\" does not have a date time field at \"%s\"", e.Name(), e.Source().Location),
				})
			}
		}

		return result
	}
}
