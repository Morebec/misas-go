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

package store

import (
	"github.com/morebec/misas-go/misas/event"
	"time"
)

type StreamDeletedEvent struct {
	StreamID  string
	Reason    *string
	DeletedAt time.Time
}

const StreamDeletedEventTypeName event.PayloadTypeName = "es.stream.deleted"

func (s StreamDeletedEvent) TypeName() event.PayloadTypeName {
	return StreamDeletedEventTypeName
}
