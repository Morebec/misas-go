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

package query

import "github.com/morebec/misas-go/misas"

// PayloadTypeName TypeName Represents the unique type name of a Query for serialization and discriminatory purposes.
type PayloadTypeName string

type Payload interface {
	TypeName() PayloadTypeName
}

// Query Represents an intent by an agent (human or machine) to get information about the system's state.
type Query struct {
	Payload  Payload
	Metadata misas.Metadata
}
