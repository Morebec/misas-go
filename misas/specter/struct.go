package specter

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
	Annotations []string `hcl:"annotations,optional"`
}

type Struct struct {
	Nam    string        `hcl:"name,label"`
	Desc   string        `hcl:"description"`
	Fields []StructField `hcl:"field,block"`
	Src    specter.Source
}

func (c *Struct) Name() specter.SpecificationName {
	return specter.SpecificationName(c.Nam)
}

func (c *Struct) Type() specter.SpecificationType {
	return "command"
}

func (c *Struct) Description() string {
	return c.Desc
}

func (c *Struct) Source() specter.Source {
	return c.Src
}

func (c *Struct) SetSource(s specter.Source) {
	c.Src = s
}

func (c *Struct) Dependencies() []specter.SpecificationName {
	var deps []specter.SpecificationName
	for _, f := range c.Fields {
		if DataType(f.Type).IsUserDefined() {
			deps = append(deps, specter.SpecificationName(f.Type))
		}
	}
	return deps
}
