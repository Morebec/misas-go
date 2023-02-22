package specter

import "github.com/morebec/specter"

type System struct {
	SName        string   `hcl:"name,label"`
	SDescription string   `hcl:"description"`
	SpecSources  []string `hcl:"sources"`
	Src          specter.Source

	Annots Annotations `hcl:"annotations,optional"`
}

func (s *System) Annotations() Annotations {
	return s.Annots
}

func (s *System) Name() specter.SpecificationName {
	return specter.SpecificationName(s.SName)
}

func (s *System) Type() specter.SpecificationType {
	return "system"
}

func (s *System) Description() string {
	return s.SDescription
}

func (s *System) Source() specter.Source {
	return s.Src
}

func (s *System) SetSource(src specter.Source) {
	s.Src = src
}

func (s *System) Dependencies() []specter.SpecificationName {
	return nil
}
