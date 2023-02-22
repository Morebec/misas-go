package specter

import "github.com/morebec/specter"

type HCLFileConfig struct {
	Constants []specter.HCLVariableConfig `hcl:"const,block"`
	Systems   []*System                   `hcl:"system,block"`
	Commands  []*Command                  `hcl:"command,block"`
	Queries   []*Query                    `hcl:"query,block"`
	Events    []*Event                    `hcl:"event,block"`
	Enums     []*Enum                     `hcl:"enum,block"`
	Structs   []*Struct                   `hcl:"struct,block"`
}

func (c HCLFileConfig) Specs() []specter.Specification {
	var grp []specter.Specification

	for _, s := range c.Systems {
		grp = append(grp, s)
	}

	for _, s := range c.Commands {
		grp = append(grp, s)
	}

	for _, s := range c.Queries {
		grp = append(grp, s)
	}

	for _, s := range c.Events {
		grp = append(grp, s)
	}

	for _, s := range c.Enums {
		grp = append(grp, s)
	}

	for _, s := range c.Structs {
		grp = append(grp, s)
	}

	return grp
}
