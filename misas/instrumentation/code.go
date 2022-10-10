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

package instrumentation

import (
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"runtime"
	"strings"
)

const (
	CodeLangKey        = attribute.Key("code.lang.name")
	CodeLangVersionKey = attribute.Key("code.lang.version")
)

const Go = "Go"

func NewStackTraceAttributes() []attribute.KeyValue {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	parts := strings.Split(frame.Function, ".")
	parts = strings.Split(parts[len(parts)-2], "/")
	packageName := parts[len(parts)-1]

	return []attribute.KeyValue{
		semconv.CodeFunctionKey.String(frame.Function),
		semconv.CodeNamespaceKey.String(packageName),
		semconv.CodeFilepathKey.String(frame.File),
		semconv.CodeLineNumberKey.Int(frame.Line),
		CodeLangKey.String(Go),
		CodeLangVersionKey.String(runtime.Version()),
	}
}
