package specter

import (
	"github.com/morebec/specter"
	"os"
)

func SpecificationTool() *specter.Specter {
	return specter.New(
		specter.WithLogger(specter.NewColoredOutputLogger(specter.ColoredOutputLoggerConfig{
			EnableColors: true,
			Writer:       os.Stdout,
		})),
		specter.WithSourceLoaders(specter.NewLocalFileSourceLoader()),
		specter.WithLoaders(specter.NewHCLFileConfigSpecLoader(func() specter.HCLFileConfig {
			return &HCLFileConfig{}
		})),
		specter.WithLinters(
			specter.SpecificationMustNotHaveUndefinedNames(),
			specter.SpecificationsMustHaveDescriptionAttribute(),
			specter.SpecificationsMustHaveLowerCaseNames(),
			specter.SpecificationsMustHaveUniqueNames(),

			EventsMustHaveDateTimeField(),
		),
		specter.WithProcessors(GoCodeGenerator{}),
		specter.WithOutputProcessors(specter.NewWriteFilesProcessor(specter.WriteFileOutputsProcessorConfig{
			UseRegistry: true,
		})),
		specter.WithExecutionMode(specter.FullMode),
	)
}
