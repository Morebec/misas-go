package spectool

import (
	"context"
	"github.com/morebec/misas-go/spectool/builtin"
	"github.com/morebec/misas-go/spectool/gogenerator"
	"github.com/morebec/misas-go/spectool/processing"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func Run(ctx context.Context, systemSpecFile string, steps ...processing.Step[*processing.Context]) error {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.DisableStacktrace = true
	config.DisableCaller = true
	logger, _ := config.Build()

	pCtx := &processing.Context{
		Context:   ctx,
		StartedAt: time.Now(),
		Logger:    logger.Sugar(),
	}

	doWork := func() error {
		pCtx.Logger.Infof("Loading system spec at %s ...", systemSpecFile)

		sources, err := processing.LoadSourcesFromFile(systemSpecFile)
		if err != nil {
			return errors.Wrapf(err, "failed loading system spec at %s", systemSpecFile)
		}

		pCtx.AddSources(sources)

		pCtx.SystemSpec, err = processing.SystemSpecDeserializer().Deserialize(sources[0])
		if err != nil {
			return errors.Wrapf(err, "failed loading system spec at %s", systemSpecFile)
		}

		pCtx.Logger.Info("System spec loaded successfully.")

		err = processing.RunSteps[*processing.Context](pCtx, steps...)
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
func LoadSpecs(deserializers ...processing.SpecDeserializer) processing.Step[*processing.Context] {
	return func(ctx *processing.Context) error {
		systemSpec := ctx.SystemSpec
		props := systemSpec.Properties.(processing.SystemSpecProperties)

		var paths []string
		paths = append(paths, filepath.Dir(systemSpec.Source.Location))
		paths = append(paths, props.Sources...)

		ctx.Logger.Infof("Loading specs from paths [%s] ...", strings.Join(paths, ","))

		sources, err := processing.LoadSourcesAtPaths(paths...)
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

		d := processing.NewCompositeDeserializer(deserializers...)
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
func LintSpecs(linters ...processing.Linter) processing.Step[*processing.Context] {
	return func(ctx *processing.Context) error {
		linter := processing.CompositeLinter(linters...)

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
func ResolveDependencies(providers ...processing.DependencyProvider) processing.Step[*processing.Context] {
	return func(ctx *processing.Context) error {
		var nodes []processing.DependencyNode

		ctx.Logger.Info("Resolving dependencies ...")
		for _, provider := range providers {
			n, err := provider(ctx.SystemSpec, ctx.Specs)
			if err != nil {
				return errors.Wrap(err, "failed resolving dependencies")
			}
			nodes = append(nodes, n...)
		}

		graph := processing.NewDependencyGraph(nodes...)
		resolvedDependencies, err := graph.Resolve()
		if err != nil {
			return errors.Wrap(err, "failed resolving dependencies")
		}

		ctx.SetResolvedDependencies(resolvedDependencies)

		ctx.Logger.Info("dependencies resolved successfully.")
		return nil
	}
}

func GenerateCommandDocumentation() processing.Step[*processing.Context] {
	return func(ctx *processing.Context) error {
		ctx.Logger.Info("Generating documentation for commands ...")

		ctx.Logger.Info("Documentation for commands generated successfully.")
		return nil
	}
}

// WriteOutputFiles returns a step that writes the output files
func WriteOutputFiles() processing.Step[*processing.Context] {
	return func(ctx *processing.Context) error {
		ctx.Logger.Info("Writing output files ...")

		registry := processing.NewOutputFileRegistry()

		for _, file := range ctx.OutputFiles {
			ctx.Logger.Infof("Writing file %s ...", file.Path)
			err := os.WriteFile(file.Path, file.Data, os.ModePerm)
			if err != nil {
				return errors.Wrapf(err, "failed writing output file at %s", file.Path)
			}
			registry.Files = append(registry.Files, file.Path)
		}

		ctx.Logger.Info("Writing output file registry ...")
		if err := registry.Write(); err != nil {
			return errors.Wrap(err, "failed writing output files")
		}
		ctx.Logger.Info("Output file registry written successfully.")

		ctx.Logger.Info("Output files written successfully.")

		return nil
	}
}

func CleanOutputFiles() processing.Step[*processing.Context] {
	return func(ctx *processing.Context) error {
		ctx.Logger.Info("Cleaning previous output files from registry ...")
		if _, err := os.Stat(processing.OutputFileRegistryFileName); errors.Is(err, os.ErrNotExist) {
			ctx.Logger.Warn("No output file registry, skipping ...")
			return nil
		}

		registry, err := processing.LoadOutputFileRegistry()
		if err != nil {
			return errors.Wrap(err, "failed cleaning output file registry")
		}
		if err := registry.Clean(); err != nil {
			return errors.Wrap(err, "failed cleaning output file registry")
		}

		ctx.Logger.Info("Previous output files cleaned from registry successfully.")
		return nil
	}
}

func Default(systemSpecFile string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return Run(
			ctx,
			systemSpecFile,
			CleanOutputFiles(),
			LoadSpecs(
				processing.SystemSpecDeserializer(),
				builtin.CommandDeserializer(),
				builtin.QueryDeserializer(),
				builtin.EventDeserializer(),
				builtin.StructDeserializer(),
				builtin.EnumDeserializer(),
				builtin.HTTPEndpointDeserializer(),
			),

			LintSpecs(
				// Common
				processing.SpecificationsMustNotHaveUndefinedTypes(),
				processing.SpecificationMustNotHaveUndefinedTypeNames(),
				processing.SpecificationsMustNotHaveDuplicateTypeNames(),
				processing.SpecificationsMustHaveDescriptions(),
				processing.SpecificationsMustHaveLowerCaseTypeNames(),

				processing.SpecificationsShouldFollowNamingConvention(),

				// Structs
				builtin.StructFieldsMustHaveTypeLinter(),
				builtin.StructFieldsShouldHaveDescriptionLinter(),

				// Enums
				builtin.EnumBaseTypeShouldBeSupportedEnumBaseType(),

				// Commands
				builtin.CommandFieldsMustHaveTypeLinter(),
				builtin.CommandFieldsShouldHaveDescriptionLinter(),

				// Queries
				builtin.QueryFieldsMustHaveTypeLinter(),
				builtin.QueryFieldsShouldHaveDescriptionLinter(),

				// Events
				builtin.EventFieldsMustHaveTypeLinter(),
				builtin.EventFieldsShouldHaveDescriptionLinter(),
				builtin.EventsMustHaveDateTimeField(),

				// HTTP Endpoints
				builtin.HTTPEndpointsShouldFollowNamingConvention(),
				builtin.HTTPEndpointPathsShouldStartWithForwardSlash(),
				builtin.HTTPEndpointPathsShouldNotEndWithForwardSlash(),
				builtin.HTTPEndpointPathsShouldBeUnique(),
				builtin.HTTPEndpointPathShouldBeLowercase(),
				builtin.HTTPEndpointsShouldHaveEitherGETorPOSTMethod(),
				builtin.HTTPEndpointsWithCommandRequestTypeMustHaveMethodPOST(),
				builtin.HTTPEndpointsWithQueryRequestTypeMustHaveMethodGET(),
				builtin.HTTPEndpointResponseShouldHaveValidStatusCode(),
			),

			ResolveDependencies(
				processing.SystemDependencyProvider(),
				builtin.CommandDependencyProvider(),
				builtin.QueryDependencyProvider(),
				builtin.EventDependencyProvider(),
				builtin.StructDependencyProvider(),
				builtin.EnumDependencyProvider(),
				builtin.HTTPEndpointDependencyProvider(),
			),

			gogenerator.GoProcessor(),

			WriteOutputFiles(),
		)
	}
}
