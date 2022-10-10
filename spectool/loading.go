package spectool

import (
	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// SpecSource represents a source file for a specification.
type SpecSource struct {
	// Absolute path to the source.
	Location string

	// Content of the Source.
	Data []byte

	// Type of the spec of the source.
	SpecType SpecType
}

type ParsedSpec struct {
	Spec
}

// loads a list of paths of files and directories and returns the encountered specs as a list of MappedSpec.
func loadSourcesAtPaths(paths ...string) ([]SpecSource, error) {
	// Use a map to avoid duplicates.
	sources := map[string]SpecSource{}

	// Use a map to avoid path duplicates.
	pathDeduplicationMap := map[string]struct{}{}

	for _, path := range paths {
		if _, found := pathDeduplicationMap[path]; found {
			continue
		}
		pathDeduplicationMap[path] = struct{}{}

		fileInfo, err := os.Stat(path)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot read file info at %s", path)
		}

		if fileInfo.IsDir() {
			loadedSources, err := loadSourcesInDirectory(path)
			if err != nil {
				return nil, errors.Wrapf(err, "failed loading specs at %s", path)
			}
			for _, s := range loadedSources {
				sources[s.Location] = s
			}
		} else {
			source, err := loadSourceFromFile(path)
			if err != nil {
				return nil, errors.Wrapf(err, "failed loading spec at %s", path)
			}
			sources[source.Location] = source
		}
	}

	return MapValues(sources), nil
}

// Loads the sources found in a directory.
func loadSourcesInDirectory(path string) ([]SpecSource, error) {
	var sources []SpecSource

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(info.Name(), ".spec.yaml") {
			return nil
		}

		src, err := loadSourceFromFile(path)
		if err != nil {
			return err
		}

		sources = append(sources, src)

		return nil
	})

	if err != nil {
		return nil, errors.Wrapf(err, "failed loading directory %s", path)
	}

	return sources, nil
}

// loads a source from a file at a given path.
func loadSourceFromFile(path string) (SpecSource, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return SpecSource{}, errors.Wrapf(err, "failed loading file %s", path)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return SpecSource{}, errors.Wrapf(err, "failed loading file %s", path)
	}

	if info.IsDir() {
		return SpecSource{}, errors.Errorf("failed loading file \"%s\", expected file got directory", path)
	}

	data, err := ioutil.ReadFile(absPath)
	if err != nil {
		return SpecSource{}, errors.Wrapf(err, "failed loading file %s", path)
	}

	parsedYaml, err := parseYaml(data)
	if err != nil {
		return SpecSource{}, errors.Wrapf(err, "failed loading file %s", path)
	}

	if !MapHasKey(parsedYaml, "version") {
		parsedYaml["version"] = ToolVersion
	}

	specVersion, err := version.NewVersion(MapGetOrDefault(parsedYaml, "version", "").(string))
	if err != nil {
		return SpecSource{}, errors.Wrapf(err, "failed loading file %s: invalid version", path)
	}
	toolVersion, err := version.NewVersion(ToolVersion)
	if err != nil {
		panic(errors.Wrap(err, "Invalid Tool Version"))
	}

	if specVersion.GreaterThan(toolVersion) {
		return SpecSource{}, errors.Errorf("failed loading file %s: invalid version %s, expected %s", path, specVersion.String(), ToolVersion)
	}

	ensureHasKeys := func(m map[string]any, keys ...string) error {
		for _, k := range keys {
			if !MapHasKey(m, k) {
				return errors.Errorf("invalid spec at %s, property %s was not defined", path, k)
			}
		}
		return nil
	}

	if err := ensureHasKeys(
		parsedYaml,
		"type",
		"typeName",
		"version",
	); err != nil {
		return SpecSource{}, errors.Wrapf(err, "failed loading file %s", path)
	}

	specType := parsedYaml["type"].(string)

	return SpecSource{Location: absPath, Data: data, SpecType: SpecType(specType)}, nil
}

// Parses YAML.
func parseYaml(content []byte) (map[string]any, error) {
	// Load as YAML.
	data := make(map[string]any)
	if err := yaml.Unmarshal(content, &data); err != nil {
		return nil, errors.Wrap(err, "failed loading spec from yaml")
	}

	return data, nil
}
