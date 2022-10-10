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

import "context"

// Handler System responsible for handling a prediction.
type Handler interface {
	Handle(p Prediction, ctx context.Context) error
}

// HandlerFunc allows defining an event handler using a function
type HandlerFunc func(p Prediction, ctx context.Context) error

func (h HandlerFunc) Handle(p Prediction, ctx context.Context) error {
	return h(p, ctx)
}
