// Copyright 2022 Morébec
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package misas

// Metadata Represents arbitrary metadata.
// All functions of metadata return a copy of the metadata.
type Metadata map[string]any

// Get a value for key or a given default value if not found.
func (m Metadata) Get(k string, d any) any {
	if m == nil {
		return nil
	}
	if v, found := m[k]; found {
		return v
	}
	return nil
}

// Set returns a copy of the metadata a given value for a key.
func (m Metadata) Set(k string, v any) Metadata {
	if m == nil {
		//goland:noinspection GoAssignmentToReceiver
		m = map[string]any{}
	}
	m[k] = v
	return m
}

// Has Indicates if this metadata contains a specific key
func (m Metadata) Has(k string) bool {
	if m == nil {
		return false
	}
	_, found := m[k]
	return found
}

// Merge Adds all the keys and values of the passed metadata to this metadata and returns the result.
// if a key is present in both the value from the passed metadata will be used.
func (m Metadata) Merge(om Metadata, overwriteKeys bool) Metadata {
	if m == nil {
		return om
	}

	for k, v := range om {
		if !m.Has(k) || overwriteKeys {
			m[k] = v
		}
	}
	return m
}
