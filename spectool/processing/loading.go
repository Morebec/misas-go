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
	var sources []spec.Source

	// Use a map to avoid path duplicates.
	deduplicatedPaths := map[string]string{}

	for _, p := range paths {
		deduplicatedPaths[p] = p
	}

	for path := range deduplicatedPaths {
		fileInfo, err := os.Stat(path)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot read file info at %s", path)
		}

		if fileInfo.IsDir() {
			loadedSources, err := loadSourcesInDirectory(path)
			if err != nil {
				return nil, errors.Wrapf(err, "failed loading specs at %s", path)
			}
			sources = append(sources, loadedSources...)
		} else {
			loadedSources, err := LoadSourcesFromFile(path)
			if err != nil {
				return nil, errors.Wrapf(err, "failed loading spec at %s", path)
			}
			sources = append(sources, loadedSources...)
		}
	}

	return sources, nil
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

		src, err := LoadSourcesFromFile(path)
		if err != nil {
			return err
		}

		sources = append(sources, src...)

		return nil
	})

	if err != nil {
		return nil, errors.Wrapf(err, "failed loading directory %s", path)
	}

	return sources, nil
}

// LoadSourcesFromFile loads a source from a file at a given path.
func LoadSourcesFromFile(path string) ([]spec.Source, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed loading file %s", path)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed loading file %s", path)
	}

	if info.IsDir() {
		return nil, errors.Errorf("failed loading file \"%s\", expected file got directory", path)
	}

	data, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed loading file %s", path)
	}

	loadedSources, err := LoadSourcesFromData(data)
	if err != nil {
		return nil, errors.Wrapf(err, "failed loading file %s", path)
	}

	var sources []spec.Source
	for _, s := range loadedSources {
		s.Location = absPath
		sources = append(sources, s)
	}

	return sources, nil
}

const MultiSpecType spec.Type = "multi"

func LoadSourcesFromData(data []byte) ([]spec.Source, error) {
	parsedYaml, err := parseYaml(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed loading spec data")
	}

	specs, err := LoadSpecFromParseYaml(parsedYaml)
	if err != nil {
		return nil, errors.Wrap(err, "failed loading spec data")
	}

	for _, s := range specs {
		if s.Data == nil {
			s.Data = data
		}
	}

	return specs, nil
}

func LoadSpecFromParseYaml(parsedYaml map[string]any) ([]spec.Source, error) {
	if !maputils.MapHasKey(parsedYaml, "version") {
		parsedYaml["specVersion"] = SpecFormatVersion
	}

	specVersion, err := version.NewVersion(maputils.MapGetOrDefault(parsedYaml, "specVersion", SpecFormatVersion).(string))
	if err != nil {
		return nil, errors.Wrapf(err, "invalid spec format version")
	}
	toolVersion, err := version.NewVersion(SpecFormatVersion)
	if err != nil {
		panic(errors.Wrap(err, "invalid spec format version"))
	}

	if specVersion.GreaterThan(toolVersion) {
		return nil, errors.Errorf(
			"invalid spec format version %s, expected %s",
			specVersion.String(),
			SpecFormatVersion,
		)
	}
	parsedYaml["specVersion"] = specVersion

	ensureHasKeys := func(m map[string]any, keys ...string) error {
		for _, k := range keys {
			if !maputils.MapHasKey(m, k) {
				return errors.Errorf("property %s was not defined", k)
			}
		}
		return nil
	}

	if err := ensureHasKeys(
		parsedYaml,
		"type",
		"specVersion",
	); err != nil {
		return nil, errors.Wrap(err, "failed loading spec data")
	}

	specType := parsedYaml["type"].(string)

	if specType == string(MultiSpecType) {
		innerSpecs, ok := parsedYaml["specs"].(map[string]any)
		if !ok {
			return nil, errors.New("invalid multi: no specs defined")
		}
		var out []spec.Source
		for _, s := range innerSpecs {
			fromParseYaml, err := LoadSpecFromParseYaml(s.(map[string]any))
			if err != nil {
				return nil, errors.Wrap(err, "failed loading spec data")
			}
			out = append(out, fromParseYaml...)
		}
		return out, nil
	}

	if err := ensureHasKeys(
		parsedYaml,
		"typeName",
	); err != nil {
		return nil, errors.Wrap(err, "failed loading spec data")
	}

	data, err := yaml.Marshal(parsedYaml)
	if err != nil {
		return nil, errors.Wrap(err, "failed loading spec data")
	}

	return []spec.Source{
		{
			Data:     data,
			SpecType: spec.Type(specType),
		},
	}, nil
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
