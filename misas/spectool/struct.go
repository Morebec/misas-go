package spectool

import "github.com/morebec/specter"

type StructField struct {
	Name        string   `hcl:"name,label"`
	Description string   `hcl:"description"`
	Type        DataType `hcl:"type"`
	Nullable    bool     `hcl:"nullable,optional"`
	Deprecation string   `hcl:"deprecation,optional"`
	Example     string   `hcl:"example,optional"`
	Default     string   `hcl:"default,optional"`
	Required    bool     `hcl:"required,optional"`

	// Annotations are used to tag a field with specific data to indicate additional information about the field.
	// One useful tag is the personal_data tag that indicates that this field contains personal information.
	Annotations Annotations `hcl:"annotations,optional"`
}

type Struct struct {
	Nam    string        `hcl:"name,label"`
	Desc   string        `hcl:"description"`
	Fields []StructField `hcl:"field,block"`
	Src    specter.Source

	Annots Annotations `hcl:"annotations,optional"`
}

func (s *Struct) Annotations() Annotations {
	return s.Annots
}

func (s *Struct) Name() specter.SpecificationName {
	return specter.SpecificationName(s.Nam)
}

func (s *Struct) Type() specter.SpecificationType {
	return "struct"
}

func (s *Struct) Description() string {
	return s.Desc
}

func (s *Struct) Source() specter.Source {
	return s.Src
}

func (s *Struct) SetSource(src specter.Source) {
	s.Src = src
}

func (s *Struct) Dependencies() []specter.SpecificationName {
	var deps []specter.SpecificationName
	for _, f := range s.Fields {
		if DataType(f.Type).IsUserDefined() {
			deps = append(deps, specter.SpecificationName(f.Type))
		}
	}
	return deps
}
