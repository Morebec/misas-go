package processing

import (
	"context"
	"github.com/morebec/misas-go/spectool/spec"
	"go.uber.org/zap"
	"time"
)

type Context struct {
	context.Context
	Logger *zap.SugaredLogger

	StartedAt time.Time
	EndedAt   time.Time

	Sources      []spec.Source
	SystemSpec   spec.Spec
	Specs        spec.Group
	Dependencies ResolvedDependencies
	OutputFiles  []OutputFile
}

func (c *Context) SetSystem(s spec.Spec) {
	if s.Type != SystemSpecType {
		panic(spec.UnexpectedSpecTypeError(s.Type, SystemSpecType))
	}

	c.SystemSpec = s
}

func (c *Context) AddSources(s []spec.Source) {
	c.Sources = append(c.Sources, s...)
}

func (c *Context) AddSpecGroup(specs spec.Group) {
	c.Specs.Merge(specs)
}

func (c *Context) AddSpec(spec spec.Spec) {
	c.Specs = append(c.Specs, spec)
}

func (c *Context) SetResolvedDependencies(graph ResolvedDependencies) {
	c.Dependencies = graph
}

func (c *Context) AddOutputFile(file OutputFile) {
	c.OutputFiles = append(c.OutputFiles, file)
}
