package syntax

import "testing"

func TestUserProfile(t *testing.T) {
	got := UserProfile("Alice", 18)
	want := "Alice is 18 years old"

	if got != want {
		t.Fatalf("UserProfile() = %q, want %q", got, want)
	}
}

func TestRectangleArea(t *testing.T) {
	got := RectangleArea(3, 4)
	want := 12.0

	if got != want {
		t.Fatalf("RectangleArea() = %v, want %v", got, want)
	}
}

func TestAverage(t *testing.T) {
	got := Average(95, 10)
	want := 9.5

	if got != want {
		t.Fatalf("Average() = %v, want %v", got, want)
	}
}

func TestIsAdult(t *testing.T) {
	cases := []struct {
		name string
		age  int
		want bool
	}{
		{name: "adult at 18", age: 18, want: true},
		{name: "adult over 18", age: 30, want: true},
		{name: "minor", age: 17, want: false},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAdult(tt.age)
			if got != tt.want {
				t.Fatalf("IsAdult(%d) = %v, want %v", tt.age, got, tt.want)
			}
		})
	}
}

func TestFormatScore(t *testing.T) {
	got := FormatScore(95)
	want := "score=95"

	if got != want {
		t.Fatalf("FormatScore() = %q, want %q", got, want)
	}
}
