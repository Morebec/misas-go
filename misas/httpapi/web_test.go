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

package httpapi

import (
	"reflect"
	"testing"
)

func TestNewSuccessResponse(t *testing.T) {
	type args struct {
		data any
	}
	tests := []struct {
		name string
		args args
		want Response
	}{
		{
			name: "new success response",
			args: args{
				data: 123,
			},
			want: Response{
				Status: Success,
				Data:   123,
				Error:  nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSuccessResponse(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSuccessResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewErrorResponse(t *testing.T) {
	type args struct {
		errorType string
		message   string
		data      any
	}
	tests := []struct {
		name string
		args args
		want Response
	}{
		{
			name: "",
			args: args{
				errorType: "unit-test-failed",
				message:   "the unit test failed",
				data:      123,
			},
			want: Response{
				Status: Failure,
				Data:   nil,
				Error: &Error{
					Type:    "unit-test-failed",
					Message: "the unit test failed",
					Data:    123,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewErrorResponse(tt.args.errorType, tt.args.message, tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewErrorResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}
