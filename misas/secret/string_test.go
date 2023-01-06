package secret

import "testing"

func TestString_String(t *testing.T) {
	tests := []struct {
		testName string
		given    String
		want     string
	}{
		{
			testName: "",
			given:    String("hello world"),
			want:     RedactedValue,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := tt.given.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString_Value(t *testing.T) {
	tests := []struct {
		testName string
		given    String
		want     string
	}{
		{
			testName: "",
			given:    String("hello world"),
			want:     "hello world",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := tt.given.Value(); got != tt.want {
				t.Errorf("Value() = %v, want %v", got, tt.want)
			}
		})
	}
}
