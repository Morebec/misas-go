package spectool

import (
	"context"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const ToolVersion = "1.0"

type ProcessingContext struct {
	context.Context
	Logger *zap.SugaredLogger

	StartedAt time.Time
	EndedAt   time.Time

	Sources      []SpecSource
	SystemSpec   Spec
	Specs        SpecGroup
	Dependencies ResolvedDependencies
	OutputFiles  []OutputFile
}

func (c *ProcessingContext) SetSystem(spec Spec) {
	if spec.Type != SystemSpecType {
		panic(UnexpectedSpecTypeError(spec.Type, SystemSpecType))
	}

	c.SystemSpec = spec
}

func (c *ProcessingContext) AddSources(s []SpecSource) {
	c.Sources = append(c.Sources, s...)
}

func (c *ProcessingContext) AddSpecGroup(specs SpecGroup) {
	c.Specs.Merge(specs)
}

func (c *ProcessingContext) AddSpec(spec Spec) {
	c.Specs = append(c.Specs, spec)
}

func (c *ProcessingContext) SetResolvedDependencies(graph ResolvedDependencies) {
	c.Dependencies = graph
}

func (c *ProcessingContext) AddOutputFile(file OutputFile) {
	c.OutputFiles = append(c.OutputFiles, file)
}

func Run(ctx context.Context, systemSpecFile string, steps ...Step[*ProcessingContext]) error {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.DisableStacktrace = true
	config.DisableCaller = true
	logger, _ := config.Build()

	pCtx := &ProcessingContext{
		Context:   ctx,
		StartedAt: time.Now(),
		Logger:    logger.Sugar(),
	}

	doWork := func() error {
		pCtx.Logger.Infof("Loading system spec at %s ...", systemSpecFile)

		source, err := loadSourceFromFile(systemSpecFile)
		if err != nil {
			return errors.Wrapf(err, "failed loading system spec at %s", systemSpecFile)
		}

		pCtx.AddSources([]SpecSource{source})

		pCtx.SystemSpec, err = SystemSpecDeserializer().Deserialize(source)
		if err != nil {
			return errors.Wrapf(err, "failed loading system spec at %s", systemSpecFile)
		}

		pCtx.Logger.Info("System spec loaded successfully.")

		err = RunSteps[*ProcessingContext](pCtx, steps...)
		if err != nil {
			return errors.Wrap(err, "spectool failed")
		}
		return nil
	}

	if err := doWork(); err != nil {
		logger.Error(err.Error())
		return err
	}

	return nil

}

// LoadSpecs returns a Step that allows loading the specs associated with the system Spec.
func LoadSpecs(deserializers ...SpecDeserializer) Step[*ProcessingContext] {
	return func(ctx *ProcessingContext) error {
		systemSpec := ctx.SystemSpec
		props := systemSpec.Properties.(SystemSpecProperties)

		var paths []string
		paths = append(paths, filepath.Dir(systemSpec.Source.Location))
		paths = append(paths, props.Sources...)

		ctx.Logger.Infof("Loading specs from paths [%s] ...", strings.Join(paths, ","))

		sources, err := loadSourcesAtPaths(paths...)
		if err != nil {
			return err
		}

		// remove system from the list of sources.
		for i, s := range sources {
			if s.Location == systemSpec.Source.Location {
				sources[i] = sources[len(sources)-1]
				sources = sources[:len(sources)-1]
			}
		}

		d := NewCompositeDeserializer(deserializers...)
		for _, s := range sources {
			spec, err := d.Deserialize(s)
			ctx.AddSpec(spec)
			if err != nil {
				return err
			}
		}

		ctx.Logger.Infof("(%d) specs loaded successfully.", len(ctx.Specs))

		return nil
	}
}

// LintSpecs returns a Step that allows linting the specs of the system for errors or warnings.
func LintSpecs(linters ...Linter) Step[*ProcessingContext] {
	return func(ctx *ProcessingContext) error {
		linter := CompositeLinter(linters...)

		ctx.Logger.Infof("Linting (%d) specs ...", len(ctx.Specs)+1) // +1 to include system spec as well.
		warnings, errs := linter(ctx.SystemSpec, ctx.Specs)
		if warnings != nil {
			ctx.Logger.Warnf("there were (%d) warnings while linting specs", len(warnings))
			for _, warn := range warnings {
				ctx.Logger.Warn(warn)
			}
		}

		if errs != nil {
			ctx.Logger.Errorf("there were (%d) errors while linting specs", len(errs))
			for _, err := range errs {
				ctx.Logger.Error(err)
			}
			return errors.Wrap(errs, "failed linting specs")
		}

		ctx.Logger.Infof("(%d) specs linted successfully.", len(ctx.Specs)+1) // +1 to include system spec as well.
		return nil
	}
}

// ResolveDependencies returns a Step that resolves the dependencies of the system.
func ResolveDependencies(providers ...DependencyProvider) Step[*ProcessingContext] {
	return func(ctx *ProcessingContext) error {
		var nodes []DependencyNode

		ctx.Logger.Info("Resolving dependencies ...")
		for _, provider := range providers {
			n, err := provider(ctx.SystemSpec, ctx.Specs)
			if err != nil {
				return errors.Wrap(err, "failed resolving dependencies")
			}
			nodes = append(nodes, n...)
		}

		graph := NewDependencyGraph(nodes...)
		resolvedDependencies, err := graph.Resolve()
		if err != nil {
			return errors.Wrap(err, "failed resolving dependencies")
		}

		ctx.SetResolvedDependencies(resolvedDependencies)

		ctx.Logger.Info("dependencies resolved successfully.")
		return nil
	}
}

func GenerateCommandDocumentation() Step[*ProcessingContext] {
	return func(ctx *ProcessingContext) error {
		ctx.Logger.Info("Generating documentation for commands ...")

		ctx.Logger.Info("Documentation for commands generated successfully.")
		return nil
	}
}

// WriteOutputFiles returns a step that writes the output files
func WriteOutputFiles() Step[*ProcessingContext] {
	return func(ctx *ProcessingContext) error {
		ctx.Logger.Info("Writing output files ...")
		for _, file := range ctx.OutputFiles {
			ctx.Logger.Infof("Writing file %s ...", file.Path)
			err := os.WriteFile(file.Path, file.Data, os.ModePerm)
			if err != nil {
				return errors.Wrapf(err, "failed writing output file at %s", file.Path)
			}
		}
		ctx.Logger.Info("Output files written successfully.")

		return nil
	}
}

func Default(systemSpecFile string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return Run(
			ctx,
			systemSpecFile,
			LoadSpecs(
				SystemSpecDeserializer(),
				CommandDeserializer(),
				QueryDeserializer(),
				EventDeserializer(),
				StructDeserializer(),
				EnumDeserializer(),
				HTTPEndpointDeserializer(),
			),

			LintSpecs(
				// Common
				SpecificationsMustNotHaveUndefinedTypes(),
				SpecificationMustNotHaveUndefinedTypeNames(),
				SpecificationsMustNotHaveDuplicateTypeNames(),
				SpecificationsMustHaveDescriptions(),
				SpecificationsMustHaveLowerCaseTypeNames(),

				SpecificationsShouldFollowNamingConvention(),

				// Structs
				StructFieldsShouldHaveDescriptionLinter(),

				// Enums
				EnumBaseTypeShouldBeSupportedEnumBaseType(),

				// Commands
				CommandFieldsShouldHaveDescriptionLinter(),

				// Queries
				QueryFieldsShouldHaveDescriptionLinter(),
				
				// Events
				EventFieldsShouldHaveDescriptionLinter(),
				EventsMustHaveDateTimeField(),

				// HTTP Endpoints
				HTTPEndpointsShouldFollowNamingConvention(),
				HTTPEndpointPathsShouldStartWithForwardSlash(),
				HTTPEndpointPathsShouldNotEndWithForwardSlash(),
				HTTPEndpointPathsShouldBeUnique(),
				HTTPEndpointPathShouldBeLowercase(),
				HTTPEndpointsShouldHaveEitherGETorPOSTMethod(),
				HTTPEndpointsWithCommandRequestTypeMustHaveMethodPOST(),
				HTTPEndpointsWithQueryRequestTypeMustHaveMethodGET(),
				HTTPEndpointResponseShouldHaveValidStatusCode(),
			),

			ResolveDependencies(
				SystemDependencyProvider(),
				CommandDependencyProvider(),
				QueryDependencyProvider(),
				EventDependencyProvider(),
				StructDependencyProvider(),
				EnumDependencyProvider(),
				HTTPEndpointDependencyProvider(),
			),

			GoProcessor(),

			WriteOutputFiles(),
		)
	}
}
