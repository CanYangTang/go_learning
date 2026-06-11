package stringutil

import "testing"

func TestIsBlank(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "empty", input: "", want: true},
		{name: "spaces", input: "   ", want: true},
		{name: "tabs and newlines", input: "\t\n", want: true},
		{name: "text", input: " Alice ", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsBlank(tt.input)
			if got != tt.want {
				t.Fatalf("IsBlank() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeSpace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "trim edges", input: "  Alice  ", want: "Alice"},
		{name: "collapse spaces", input: "Alice   Bob", want: "Alice Bob"},
		{name: "tabs and newlines", input: "Alice\t\nBob", want: "Alice Bob"},
		{name: "blank", input: "   ", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeSpace(tt.input)
			if got != tt.want {
				t.Fatalf("NormalizeSpace() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestJoinNonEmpty(t *testing.T) {
	tests := []struct {
		name  string
		sep   string
		parts []string
		want  string
	}{
		{name: "comma", sep: ",", parts: []string{"a", "", "b"}, want: "a,b"},
		{name: "hyphen", sep: "-", parts: []string{"go", "", "lang"}, want: "go-lang"},
		{name: "normalizes parts", sep: " ", parts: []string{"  Alice  ", "   Bob"}, want: "Alice Bob"},
		{name: "all blank", sep: ",", parts: []string{"", "   "}, want: ""},
		{name: "no parts", sep: ",", parts: nil, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := JoinNonEmpty(tt.sep, tt.parts...)
			if got != tt.want {
				t.Fatalf("JoinNonEmpty() = %q, want %q", got, tt.want)
			}
		})
	}
}
