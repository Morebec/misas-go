package spectool

import "github.com/morebec/specter"

type EnumValue struct {
	Name  string `hcl:"name,label"`
	Value any    `hcl:"value"`
}

type Enum struct {
	Nam      string      `hcl:"name,label"`
	Desc     string      `hcl:"description"`
	Values   []EnumValue `hcl:"value,block"`
	BaseType DataType    `hcl:"type"`
	Src      specter.Source
	Annots   Annotations `hcl:"annotations,optional"`
	Meta     Metadata    `hcl:"meta,block"`
}

func (e *Enum) Metadata() Metadata {
	return e.Meta
}

func (e *Enum) Annotations() Annotations {
	return e.Annots
}

func (e *Enum) Name() specter.SpecificationName {
	return specter.SpecificationName(e.Nam)
}

func (e *Enum) Type() specter.SpecificationType {
	return "enum"
}

func (e *Enum) Description() string {
	return e.Desc
}

func (e *Enum) Source() specter.Source {
	return e.Src
}

func (e *Enum) SetSource(s specter.Source) {
	e.Src = s
}

func (e *Enum) Dependencies() []specter.SpecificationName {
	return []specter.SpecificationName{specter.SpecificationName(e.BaseType)}
}
