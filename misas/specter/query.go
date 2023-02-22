package specter

import "github.com/morebec/specter"

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

	Annots Annotations `hcl:"annotations,optional"`
}

func (q *Query) Annotations() Annotations {
	return q.Annots
}

func (q *Query) Name() specter.SpecificationName {
	return specter.SpecificationName(q.Nam)
}

func (q *Query) Type() specter.SpecificationType {
	return "query"
}

func (q *Query) Description() string {
	return q.Desc
}

func (q *Query) Source() specter.Source {
	return q.Src
}

func (q *Query) SetSource(s specter.Source) {
	q.Src = s
}

func (q *Query) Dependencies() []specter.SpecificationName {
	var deps []specter.SpecificationName
	for _, f := range q.Fields {
		if f.Type.IsUserDefined() {
			deps = append(deps, specter.SpecificationName(f.Type))
		}
	}
	return deps
}
