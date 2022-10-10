package spectool

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

// LintingErrors represents a set of errors that were encountered during the linting process.
// If a linter considers that a file is absolutely incorrect it should return it as a LintingError.
// This will cause the tool to stop and require the user to fix the situation.
// As a naming convention, use "Must" to indicate that a linter will return an error for violations,
// and use "Should" to indicate that a linter will return warnings for violations.
type LintingErrors []error

func (le LintingErrors) Error() string {
	message := fmt.Sprintf("there were linting errors (%d):\n", len(le))
	for _, e := range le {
		message += fmt.Sprintf("> %s\n", e.Error())
	}
	return message
}

// LintingWarnings represents a set of warnings to be communicated to the user but that should not halt the tool.
type LintingWarnings []string

// Linter represents a function responsible for linting specs.
type Linter func(system Spec, specs SpecGroup) (LintingWarnings, LintingErrors)

// CompositeLinter A Composite linter is responsible for running multiple linters as one.
func CompositeLinter(linters ...Linter) Linter {
	return func(system Spec, specs SpecGroup) (LintingWarnings, LintingErrors) {
		var lintingErrors LintingErrors
		var lintingWarnings LintingWarnings
		for _, linter := range linters {
			warns, errs := linter(system, specs)
			if warns != nil {
				lintingWarnings = append(lintingWarnings, warns...)
			}
			if errs != nil {
				lintingErrors = append(lintingErrors, errs...)
			}
		}
		return lintingWarnings, lintingErrors
	}
}

// SpecificationsMustNotHaveUndefinedTypes ensures that no spec has an undefined type
func SpecificationsMustNotHaveUndefinedTypes() Linter {
	return func(system Spec, specs SpecGroup) (LintingWarnings, LintingErrors) {
		var lintingErrors LintingErrors
		var lintingWarnings LintingWarnings

		for _, s := range specs {
			if s.Type == UndefinedSpecificationType {
				lintingErrors = append(lintingErrors, errors.Errorf("spec at \"%s\" has an undefined type", s.Source.Location))
			}
		}

		return lintingWarnings, lintingErrors
	}
}

// SpecificationMustNotHaveUndefinedTypeNames ensures that no spec has an undefined type name
func SpecificationMustNotHaveUndefinedTypeNames() Linter {
	return func(system Spec, specs SpecGroup) (LintingWarnings, LintingErrors) {
		var lintingErrors LintingErrors
		var lintingWarnings LintingWarnings

		for _, s := range specs {
			if s.TypeName == UndefinedSpecificationTypename {
				lintingErrors = append(lintingErrors, errors.Errorf("spec at \"%s\" has an undefined type name", s.Source.Location))
			}
		}

		return lintingWarnings, lintingErrors
	}
}

// SpecificationsMustNotHaveDuplicateTypeNames ensures that type names are unique amongst specs.
func SpecificationsMustNotHaveDuplicateTypeNames() Linter {
	return func(system Spec, specs SpecGroup) (LintingWarnings, LintingErrors) {
		var lintingErrors LintingErrors
		var lintingWarnings LintingWarnings

		// Where key is the type name and the array contains all the spec file locations where it was encountered.
		encounteredTypeNames := map[SpecTypeName][]string{}

		for _, s := range specs {
			if _, found := encounteredTypeNames[s.TypeName]; found {
				encounteredTypeNames[s.TypeName] = append(encounteredTypeNames[s.TypeName], s.Source.Location)
			} else {
				encounteredTypeNames[s.TypeName] = []string{s.Source.Location}
			}
		}

		for typeName, files := range encounteredTypeNames {
			if len(files) > 1 {
				lintingErrors = append(lintingErrors, fmt.Errorf("duplicate type name detected for \"%s\" in the following files: %s", typeName, strings.Join(files, ", ")))
			}
		}

		return lintingWarnings, lintingErrors
	}
}

// SpecificationsMustHaveDescriptions ensures that all specs have a description.
func SpecificationsMustHaveDescriptions() Linter {
	return func(system Spec, specs SpecGroup) (LintingWarnings, LintingErrors) {
		var lintingErrors LintingErrors
		var lintingWarnings LintingWarnings
		for _, s := range specs {
			if s.Description == "" {
				lintingErrors = append(lintingErrors, errors.Errorf("spec at location %s does not have a description.", s.Source.Location))
			}
		}
		return lintingWarnings, lintingErrors
	}
}

// SpecificationsMustHaveLowerCaseTypeNames ensures that all spec type names are lower case.
func SpecificationsMustHaveLowerCaseTypeNames() Linter {
	return func(system Spec, specs SpecGroup) (LintingWarnings, LintingErrors) {
		var lintingErrors LintingErrors
		var lintingWarnings LintingWarnings
		for _, s := range specs {
			if strings.ToLower(string(s.TypeName)) != string(s.TypeName) {
				lintingErrors = append(lintingErrors, errors.Errorf("spec type names must be lowercase got %s at location %s.", s.TypeName, s.Source.Location))
			}
		}
		return lintingWarnings, lintingErrors
	}
}

func SpecificationsShouldFollowNamingConvention() Linter {
	return func(system Spec, specs SpecGroup) (LintingWarnings, LintingErrors) {
		var lintingWarnings LintingWarnings
		for _, s := range specs {
			nbParts := len(s.TypeName.Parts())
			if nbParts < 3 {
				lintingWarnings = append(lintingWarnings, fmt.Sprintf("%s %s does not follow the <module>.<entity>.<sub_entity> naming convention at %s", s.Type, s.TypeName, s.Source.Location))
			}
		}

		return lintingWarnings, nil
	}
}
