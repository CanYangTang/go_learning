package mathutil

import "testing"

func TestDouble(t *testing.T) {
	got := Double(3)
	want := 6

	if got != want {
		t.Fatalf("Double() = %v, want %v", got, want)
	}
}

func TestIsEven(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want bool
	}{
		{name: "even", n: 4, want: true},
		{name: "odd", n: 5, want: false},
		{name: "zero", n: 0, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsEven(tt.n)
			if got != tt.want {
				t.Fatalf("IsEven() = %v, want %v", got, tt.want)
			}
		})
	}
}
