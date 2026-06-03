package controlflow

import (
	"reflect"
	"testing"
)

func TestGrade(t *testing.T) {
	cases := []struct {
		name  string
		score int
		want  string
	}{
		{name: "invalid low", score: -1, want: "Invalid"},
		{name: "invalid high", score: 101, want: "Invalid"},
		{name: "A lower bound", score: 90, want: "A"},
		{name: "A high", score: 100, want: "A"},
		{name: "B lower bound", score: 80, want: "B"},
		{name: "C lower bound", score: 60, want: "C"},
		{name: "D lower bound", score: 0, want: "D"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := Grade(tt.score)
			if got != tt.want {
				t.Fatalf("Grade(%d) = %q, want %q", tt.score, got, tt.want)
			}
		})
	}
}

func TestDayType(t *testing.T) {
	cases := []struct {
		day  string
		want string
	}{
		{day: "Mon", want: "weekday"},
		{day: "Fri", want: "weekday"},
		{day: "Sat", want: "weekend"},
		{day: "Sun", want: "weekend"},
		{day: "Holiday", want: "unknown"},
	}

	for _, tt := range cases {
		t.Run(tt.day, func(t *testing.T) {
			got := DayType(tt.day)
			if got != tt.want {
				t.Fatalf("DayType(%q) = %q, want %q", tt.day, got, tt.want)
			}
		})
	}
}

func TestEvenNumbers(t *testing.T) {
	got := EvenNumbers(8)
	want := []int{2, 4, 6, 8}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("EvenNumbers(8) = %#v, want %#v", got, want)
	}
}

func TestEvenNumbersNoResult(t *testing.T) {
	got := EvenNumbers(1)
	if len(got) != 0 {
		t.Fatalf("EvenNumbers(1) = %#v, want empty slice", got)
	}
}

func TestSumTo(t *testing.T) {
	cases := []struct {
		n    int
		want int
	}{
		{n: 1, want: 1},
		{n: 5, want: 15},
		{n: 10, want: 55},
	}

	for _, tt := range cases {
		got := SumTo(tt.n)
		if got != tt.want {
			t.Fatalf("SumTo(%d) = %d, want %d", tt.n, got, tt.want)
		}
	}
}

func TestMultiplicationTable(t *testing.T) {
	got := MultiplicationTable()

	if len(got) != 9 {
		t.Fatalf("MultiplicationTable() returned %d rows, want 9", len(got))
	}

	if got[0] != "1*1=1" {
		t.Fatalf("first row = %q, want %q", got[0], "1*1=1")
	}

	wantLast := "1*9=9 2*9=18 3*9=27 4*9=36 5*9=45 6*9=54 7*9=63 8*9=72 9*9=81"
	if got[8] != wantLast {
		t.Fatalf("last row = %q, want %q", got[8], wantLast)
	}
}
