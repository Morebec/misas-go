package processing

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

const OutputFileRegistryFileName = ".specs.json"

// OutputFile represents a file the needs to be outputted.
type OutputFile struct {
	// Absolute path to the file
	Path string

	// Content that should be written.
	Data []byte
}

type OutputFileRegistry struct {
	GeneratedAt time.Time `json:"generatedAt"`
	Files       []string  `json:"files"`
}

func NewOutputFileRegistry() OutputFileRegistry {
	return OutputFileRegistry{
		GeneratedAt: time.Now(),
		Files:       nil,
	}
}

func LoadOutputFileRegistry() (OutputFileRegistry, error) {
	r := OutputFileRegistry{}
	bytes, err := os.ReadFile(OutputFileRegistryFileName)
	if err != nil {
		return OutputFileRegistry{}, errors.Wrap(err, "failed loading output file registry")
	}
	if err := json.Unmarshal(bytes, &r); err != nil {
		return OutputFileRegistry{}, errors.Wrap(err, "failed loading output file registry")
	}

	return r, nil
}

func (r OutputFileRegistry) Write() error {
	// Generate a JSON file containing all output files for clean up later on
	js, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed generating output file registry")
	}
	if err := ioutil.WriteFile(OutputFileRegistryFileName, js, os.ModePerm); err != nil {
		return errors.Wrap(err, "failed generating output file registry")
	}

	return nil
}

func (r OutputFileRegistry) Clean() error {
	var wg sync.WaitGroup
	for _, f := range r.Files {
		wg.Add(1)
		f := f
		go func() {
			defer wg.Done()
			if err := os.Remove(f); err != nil {
				panic(errors.Wrap(err, "failed cleaning output registry files"))
			}
		}()
	}
	wg.Wait()

	return nil
}
