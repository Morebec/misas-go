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

// This package contains an interface to abstract the provisioning of the current time to the system.
// This interface is used to explicitly indicate that obtaining the current date and time is an infrastructural component,
// and that it should be not be accessed directly. This has the benefit of allowing one to manipulate at will how the system receives time
// and vary this provisioning based on the system's environment for example.
// The clock package proposes 3 implementations out of the box:
// - `UTCClock` which is responsible for providing the current date and time of the system in the UTC time zone.
// - `FixedClock` which always returns a certain predefined date and time.
// - `OffsetClock` which returns the date and time of the system with a given offset.
