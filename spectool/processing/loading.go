package processing

import (
	"github.com/hashicorp/go-version"
	"github.com/morebec/misas-go/spectool/maputils"
	"github.com/morebec/misas-go/spectool/spec"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const SpecFormatVersion = "1.0"

type ParsedSpec struct {
	spec.Spec
}

// LoadSourcesAtPaths loads a list of paths of files and directories and returns the encountered specs as a list of spec.Source.
func LoadSourcesAtPaths(paths ...string) ([]spec.Source, error) {
	// Use a map to avoid duplicates.
	sources := map[string]spec.Source{}

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
			source, err := LoadSourceFromFile(path)
			if err != nil {
				return nil, errors.Wrapf(err, "failed loading spec at %s", path)
			}
			sources[source.Location] = source
		}
	}

	return maputils.MapValues(sources), nil
}

// Loads the sources found in a directory.
func loadSourcesInDirectory(path string) ([]spec.Source, error) {
	var sources []spec.Source

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(info.Name(), ".spec.yaml") {
			return nil
		}

		src, err := LoadSourceFromFile(path)
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

// LoadSourceFromFile loads a source from a file at a given path.
func LoadSourceFromFile(path string) (spec.Source, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return spec.Source{}, errors.Wrapf(err, "failed loading file %s", path)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return spec.Source{}, errors.Wrapf(err, "failed loading file %s", path)
	}

	if info.IsDir() {
		return spec.Source{}, errors.Errorf("failed loading file \"%s\", expected file got directory", path)
	}

	data, err := ioutil.ReadFile(absPath)
	if err != nil {
		return spec.Source{}, errors.Wrapf(err, "failed loading file %s", path)
	}

	parsedYaml, err := parseYaml(data)
	if err != nil {
		return spec.Source{}, errors.Wrapf(err, "failed loading file %s", path)
	}

	if !maputils.MapHasKey(parsedYaml, "version") {
		parsedYaml["version"] = SpecFormatVersion
	}

	specVersion, err := version.NewVersion(maputils.MapGetOrDefault(parsedYaml, "version", "").(string))
	if err != nil {
		return spec.Source{}, errors.Wrapf(err, "failed loading file %s: invalid version", path)
	}
	toolVersion, err := version.NewVersion(SpecFormatVersion)
	if err != nil {
		panic(errors.Wrap(err, "Invalid Tool Version"))
	}

	if specVersion.GreaterThan(toolVersion) {
		return spec.Source{}, errors.Errorf("failed loading file %s: invalid version %s, expected %s", path, specVersion.String(), SpecFormatVersion)
	}

	ensureHasKeys := func(m map[string]any, keys ...string) error {
		for _, k := range keys {
			if !maputils.MapHasKey(m, k) {
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
		return spec.Source{}, errors.Wrapf(err, "failed loading file %s", path)
	}

	specType := parsedYaml["type"].(string)

	return spec.Source{Location: absPath, Data: data, SpecType: spec.Type(specType)}, nil
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
