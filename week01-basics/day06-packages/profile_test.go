package packages

import "testing"

func TestNormalizeName(t *testing.T) {
	got := NormalizeName("  Alice  ")
	want := "Alice"

	if got != want {
		t.Fatalf("NormalizeName() = %q, want %q", got, want)
	}
}

func TestUserLabel(t *testing.T) {
	got := UserLabel("  Alice  ", 18)
	want := "Alice(18)"

	if got != want {
		t.Fatalf("UserLabel() = %q, want %q", got, want)
	}
}

func TestDoubleAge(t *testing.T) {
	got := DoubleAge(18)
	want := 36

	if got != want {
		t.Fatalf("DoubleAge() = %v, want %v", got, want)
	}
}

func TestIsAdultAgeEven(t *testing.T) {
	tests := []struct {
		name string
		age  int
		want bool
	}{
		{name: "adult even", age: 18, want: true},
		{name: "adult odd", age: 19, want: false},
		{name: "minor even", age: 16, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAdultAgeEven(tt.age)
			if got != tt.want {
				t.Fatalf("IsAdultAgeEven() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPackageStatus(t *testing.T) {
	got := PackageStatus()
	want := "ready"

	if got != want {
		t.Fatalf("PackageStatus() = %q, want %q", got, want)
	}
}
