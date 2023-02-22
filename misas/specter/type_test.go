package specter

import (
	"reflect"
	"testing"
)

func TestFieldType_BaseType(t *testing.T) {
	tests := []struct {
		name string
		f    DataType
		want DataType
	}{
		{
			name: "simple base type",
			f:    String,
			want: String,
		},
		{
			name: "array of strings",
			f:    "[]" + String,
			want: String,
		},
		{
			name: "array should return array",
			f:    "[]Type",
			want: "Type",
		},
		{
			name: "map should return the type of values",
			f:    "map[string]Type",
			want: "Type",
		},
		{
			name: "array of container should return array (the left-most type)",
			f:    "[]map[string]Type",
			want: "Type",
		},
		{
			name: "map of container should return map (the left-most type)",
			f:    "map[string][]Type",
			want: "Type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.BaseType(); got != tt.want {
				t.Errorf("BaseType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldType_IsArray(t *testing.T) {
	tests := []struct {
		name string
		f    DataType
		want bool
	}{
		{
			name: "non array should return false",
			f:    String,
			want: false,
		},
		{
			name: "array should return true",
			f:    "[]Type",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.IsArray(); got != tt.want {
				t.Errorf("IsArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldType_IsContainer(t *testing.T) {
	tests := []struct {
		name string
		f    DataType
		want bool
	}{
		{
			name: "array should return true",
			f:    "[]Type",
			want: true,
		},
		{
			name: "map should return true",
			f:    "map[string]ok",
			want: true,
		},
		{
			name: "non container should return false",
			f:    String,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.IsContainer(); got != tt.want {
				t.Errorf("IsContainer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldType_IsMap(t *testing.T) {
	tests := []struct {
		name string
		f    DataType
		want bool
	}{
		{
			name: "map should return true",
			f:    "map[string]Type",
			want: true,
		},
		{
			name: "map word only should return false",
			f:    "map",
			want: false,
		},
		{
			name: "malformed map should return false",
			f:    "map[]",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.IsMap(); got != tt.want {
				t.Errorf("IsMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldType_MapInfo(t *testing.T) {
	tests := []struct {
		name string
		f    DataType
		want ContainerTypeInfo
	}{
		{
			name: "simple key value should return type of values",
			f:    "map[string]Type",
			want: ContainerTypeInfo{
				KeyType:   String,
				ValueType: "Type",
			},
		},
		{
			name: "values as array of Type should return Type",
			f:    "map[string][]Type",
			want: ContainerTypeInfo{
				KeyType:   String,
				ValueType: "[]Type",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.MapInfo(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContainerTypeInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldType_ArrayInfo(t *testing.T) {
	tests := []struct {
		name string
		f    DataType
		want ContainerTypeInfo
	}{
		{
			name: "array should return int as KeyType and values of type as ValueType",
			f:    "[]Type",
			want: ContainerTypeInfo{
				KeyType:   Int,
				ValueType: "Type",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.ArrayInfo(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ArrayInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldType_IsUserDefined(t *testing.T) {
	tests := []struct {
		name string
		f    DataType
		want bool
	}{
		{
			name: "built in types should return false",
			f:    Int,
			want: false,
		},
		{
			name: "user defined types should return true",
			f:    "user_defined",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.IsUserDefined(); got != tt.want {
				t.Errorf("IsUserDefined() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldType_IsBuiltIn(t *testing.T) {
	tests := []struct {
		name string
		f    DataType
		want bool
	}{
		{
			name: "built in types should return true",
			f:    Int,
			want: true,
		},
		{
			// even if the valueType is user defined.
			name: "container types should return true",
			f:    "map[string]user_defined",
			want: true,
		},
		{
			// even if the valueType is user defined.
			name: "container types should return true",
			f:    "[]user_defined",
			want: true,
		},
		{
			name: "user defined types should return false",
			f:    "user_defined",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.IsBuiltIn(); got != tt.want {
				t.Errorf("IsBuiltIn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldType_ExtractUserDefined(t *testing.T) {
	tests := []struct {
		name string
		f    DataType
		want DataType
	}{
		{
			name: "complex container with user defined type should return user defined",
			f:    "[]map[string][]user_defined",
			want: "user_defined",
		},
		{
			name: "complex container with builtin type should return empty string",
			f:    "[]map[string][]string",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.ExtractUserDefined(); got != tt.want {
				t.Errorf("ExtractUserDefined() = %v, want %v", got, tt.want)
			}
		})
	}
}
