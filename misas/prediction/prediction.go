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

package prediction

import (
	"time"
)

// TypeName Represents the unique type name of a Prediction for serialization and discriminatory purposes.
type TypeName string

// ID represents the unique Identifier of a prediction. This is used so the domain logic can refer to specific predictions
// in order to cancel them in the future.
type ID string

// Prediction represents an expectation for something to happen in the future. They can be seen as future events, however contrary
// to them, they can be rejected if the expectations are no longer true on the given time.
type Prediction interface {
	ID() ID

	TypeName() TypeName

	// WillOccurAt Returns the date and time at which this prediction is supposed to occur.
	WillOccurAt() time.Time
}
