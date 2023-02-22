package misas

import (
	"github.com/morebec/misas-go/specter"
)

type QueryField struct {
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

type Query struct {
	Nam    string       `hcl:"name,label"`
	Desc   string       `hcl:"description"`
	Fields []QueryField `hcl:"field,block"`
	Src    specter.Source
}

func (query *Query) Name() specter.Name {
	return specter.Name(query.Nam)
}

func (query *Query) Type() specter.Type {
	return "query"
}

func (query *Query) Description() string {
	return query.Desc
}

func (query *Query) Source() specter.Source {
	return query.Src
}

func (query *Query) SetSource(s specter.Source) {
	query.Src = s
}

func (query *Query) Dependencies() []specter.Name {
	var deps []specter.Name
	for _, f := range query.Fields {
		if DataType(f.Type).IsUserDefined() {
			deps = append(deps, specter.Name(f.Type))
		}
	}
	return deps
}
