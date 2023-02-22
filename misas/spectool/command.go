package spectool

import "github.com/morebec/specter"

type CommandField struct {
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

type Command struct {
	Nam    string         `hcl:"name,label"`
	Desc   string         `hcl:"description"`
	Fields []CommandField `hcl:"field,block"`
	Annots Annotations    `hcl:"annotations,optional"`
	Src    specter.Source
}

func (c *Command) Annotations() Annotations {
	return c.Annots
}

func (c *Command) Name() specter.SpecificationName {
	return specter.SpecificationName(c.Nam)
}

func (c *Command) Type() specter.SpecificationType {
	return "command"
}

func (c *Command) Description() string {
	return c.Desc
}

func (c *Command) Source() specter.Source {
	return c.Src
}

func (c *Command) SetSource(s specter.Source) {
	c.Src = s
}

func (c *Command) Dependencies() []specter.SpecificationName {
	var deps []specter.SpecificationName
	for _, f := range c.Fields {
		if DataType(f.Type).IsUserDefined() {
			deps = append(deps, specter.SpecificationName(f.Type))
		}
	}
	return deps
}
