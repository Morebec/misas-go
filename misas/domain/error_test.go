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

package domain

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		typeName ErrorTypeName
		opts     []ErrorOption
	}
	cause := errors.New("there was an error")

	tests := []struct {
		name string
		args args
		want Error
	}{
		{
			name: "new",
			args: args{
				typeName: "unit_test_error",
				opts: []ErrorOption{
					WithMessage("there was an error in the unit test"),
					WithData(map[string]any{
						"id": "unit_test_1",
					}),
				},
			},
			want: Error{
				message:  "there was an error in the unit test",
				typeName: "unit_test_error",
				cause:    nil,
				data: map[string]any{
					"id": "unit_test_1",
				},
			},
		},
		{
			name: "new with cause",
			args: args{
				typeName: "unit_test_error",
				opts: []ErrorOption{
					WithMessage("there was an error in the unit test"),
					WithCause(cause),
				},
			},
			want: Error{
				message:  "there was an error in the unit test",
				typeName: "unit_test_error",
				cause:    cause,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewError(tt.args.typeName, tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDomainError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "with actual domain error",
			args: args{
				err: NewError(
					"unit_test_error",
					WithMessage("there was an error in the unit test"),
				),
			},
			want: true,
		},
		{
			name: "without a domain error",
			args: args{
				err: errors.New("unit test error"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDomainError(tt.args.err); got != tt.want {
				t.Errorf("IsDomainError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDomainErrorWithTypeName(t *testing.T) {
	type args struct {
		err error
		t   ErrorTypeName
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "with actual domain error",
			args: args{
				err: NewError(
					"unit_test_error",
					WithMessage("there was an error in the unit test"),
				),
				t: "unit_test_error",
			},
			want: true,
		},
		{
			name: "with domain error but wrong type",
			args: args{
				err: NewError(
					"unit_test_error",
					WithMessage("there was an error in the unit test"),
				),
				t: "wrong_type",
			},
			want: false,
		},
		{
			name: "without a domain error",
			args: args{
				err: errors.New("unit test error"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDomainErrorWithTypeName(tt.args.err, tt.args.t); got != tt.want {
				t.Errorf("IsDomainErrorWithTypeName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError_HasTag(t *testing.T) {
	type fields struct {
		message  string
		typeName ErrorTypeName
		cause    error
		data     map[string]any
		tags     []string
	}
	type args struct {
		tag string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "with tag should return true",
			fields: fields{
				message:  "",
				typeName: "",
				cause:    nil,
				data:     nil,
				tags: []string{
					NotFoundTag,
				},
			},
			args: args{
				tag: NotFoundTag,
			},
			want: true,
		},
		{
			name: "without tag should return false",
			fields: fields{
				message:  "",
				typeName: "",
				cause:    nil,
				data:     nil,
				tags:     nil,
			},
			args: args{
				tag: NotFoundTag,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Error{
				message:  tt.fields.message,
				typeName: tt.fields.typeName,
				cause:    tt.fields.cause,
				data:     tt.fields.data,
				tags:     tt.fields.tags,
			}
			assert.Equalf(t, tt.want, d.HasTag(tt.args.tag), "HasTag(%v)", tt.args.tag)
		})
	}
}

func TestIsNotFoundDomainError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "with tag not found should return true",
			args: args{
				err: NewError(
					"unit_test_error",
					WithMessage("there was an error in the unit test"),
					WithTags(NotFoundTag),
				),
			},
			want: true,
		},
		{
			name: "without tag not found should return false",
			args: args{
				err: NewError(
					"unit_test_error",
					WithMessage("there was an error in the unit test"),
				),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsNotFoundError(tt.args.err), "IsNotFoundError(%v)", tt.args.err)
		})
	}
}

func TestWithMessagef(t *testing.T) {
	err := NewError(
		"unit_test_error",
		WithMessagef("unit test %s succeeded", "WithMessagef"),
	)
	assert.Equal(t, "unit test WithMessagef succeeded", err.message)
}

func TestWithMessage(t *testing.T) {
	err := NewError(
		"unit_test_error",
		WithMessage("unit test succeeded"),
	)
	assert.Equal(t, "unit test succeeded", err.message)
}
