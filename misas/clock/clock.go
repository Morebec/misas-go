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

package clock

import "time"

type Clock interface {
	Now() time.Time
}

// UTCClock Implementation of a Clock that returns the current time of the system as UTC.
type UTCClock struct {
}

func NewUTCClock() *UTCClock {
	return &UTCClock{}
}

func (s UTCClock) Now() time.Time {
	return time.Now().UTC()
}

// FixedClock Implementation of a Clock that always returns a Fixed date.
type FixedClock struct {
	CurrentDate time.Time
}

func NewFixedClock(currentDate time.Time) *FixedClock {
	return &FixedClock{CurrentDate: currentDate}
}

func (f FixedClock) Now() time.Time {
	return f.CurrentDate
}

// OffsetClock implementation of a Clock that returns a date with a predefined offset.
type OffsetClock struct {
	Offset time.Duration
}

func NewOffsetClock(offset time.Duration) *OffsetClock {
	return &OffsetClock{Offset: offset}
}

func (o OffsetClock) Now() time.Time {
	return time.Now().Add(o.Offset)
}
