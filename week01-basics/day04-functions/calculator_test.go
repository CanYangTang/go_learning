package functions

import "testing"

func TestAdd(t *testing.T) {
	got := Add(1.5, 2.5)
	want := 4.0

	if got != want {
		t.Fatalf("Add() = %v, want %v", got, want)
	}
}

func TestSubtract(t *testing.T) {
	got := Subtract(5, 2)
	want := 3.0

	if got != want {
		t.Fatalf("Subtract() = %v, want %v", got, want)
	}
}

func TestMultiply(t *testing.T) {
	got := Multiply(3, 4)
	want := 12.0

	if got != want {
		t.Fatalf("Multiply() = %v, want %v", got, want)
	}
}

func TestDivide(t *testing.T) {
	got, ok := Divide(10, 2)
	want := 5.0

	if !ok {
		t.Fatal("Divide() ok = false, want true")
	}

	if got != want {
		t.Fatalf("Divide() = %v, want %v", got, want)
	}
}

func TestDivideByZero(t *testing.T) {
	got, ok := Divide(10, 0)

	if ok {
		t.Fatal("Divide(10, 0) ok = true, want false")
	}

	if got != 0 {
		t.Fatalf("Divide(10, 0) = %v, want 0", got)
	}
}

func TestSum(t *testing.T) {
	got := Sum(1, 2, 3, 4)
	want := 10.0

	if got != want {
		t.Fatalf("Sum() = %v, want %v", got, want)
	}
}

func TestSumEmpty(t *testing.T) {
	got := Sum()
	want := 0.0

	if got != want {
		t.Fatalf("Sum() = %v, want %v", got, want)
	}
}

func TestApply(t *testing.T) {
	got := Apply(2, 3, Add)
	want := 5.0

	if got != want {
		t.Fatalf("Apply() = %v, want %v", got, want)
	}
}

func TestApplyWithAnonymousFunction(t *testing.T) {
	got := Apply(2, 3, func(a, b float64) float64 {
		return a*b + 1
	})
	want := 7.0

	if got != want {
		t.Fatalf("Apply() = %v, want %v", got, want)
	}
}

func TestNewCounter(t *testing.T) {
	counter := NewCounter()

	if counter() != 1 {
		t.Fatal("first counter call should return 1")
	}

	if counter() != 2 {
		t.Fatal("second counter call should return 2")
	}

	anotherCounter := NewCounter()
	if anotherCounter() != 1 {
		t.Fatal("new counter should start from 1")
	}
}
