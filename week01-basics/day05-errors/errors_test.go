package errors

import (
	stderrors "errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestDivide(t *testing.T) {
	got, err := Divide(10, 2)
	want := 5.0

	if err != nil {
		t.Fatalf("Divide() error = %v, want nil", err)
	}

	if got != want {
		t.Fatalf("Divide() = %v, want %v", got, want)
	}
}

func TestDivideByZero(t *testing.T) {
	got, err := Divide(10, 0)

	if !stderrors.Is(err, ErrDivideByZero) {
		t.Fatalf("Divide(10, 0) error = %v, want ErrDivideByZero", err)
	}

	if got != 0 {
		t.Fatalf("Divide(10, 0) = %v, want 0", got)
	}
}

func TestCalculateAddSubtractMultiply(t *testing.T) {
	tests := []struct {
		name string
		a    float64
		b    float64
		op   string
		want float64
	}{
		{name: "add", a: 2, b: 3, op: "+", want: 5},
		{name: "subtract", a: 5, b: 3, op: "-", want: 2},
		{name: "multiply", a: 4, b: 3, op: "*", want: 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Calculate(tt.a, tt.b, tt.op)
			if err != nil {
				t.Fatalf("Calculate() error = %v, want nil", err)
			}

			if got != tt.want {
				t.Fatalf("Calculate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateDivideByZeroWrapsError(t *testing.T) {
	got, err := Calculate(10, 0, "/")

	if got != 0 {
		t.Fatalf("Calculate(10, 0, /) = %v, want 0", got)
	}

	if !IsDivideByZero(err) {
		t.Fatalf("IsDivideByZero(%v) = false, want true", err)
	}

	if err == ErrDivideByZero {
		t.Fatal("Calculate() should wrap ErrDivideByZero instead of returning it directly")
	}
}

func TestCalculateInvalidOperator(t *testing.T) {
	_, err := Calculate(1, 2, "%")

	if err == nil {
		t.Fatal("Calculate() error = nil, want invalid operator error")
	}

	if !strings.Contains(err.Error(), "invalid operator") {
		t.Fatalf("Calculate() error = %v, want invalid operator message", err)
	}
}

func TestIsDivideByZeroWithWrappedError(t *testing.T) {
	err := fmt.Errorf("calculate failed: %w", ErrDivideByZero)

	if !IsDivideByZero(err) {
		t.Fatal("IsDivideByZero() = false, want true")
	}
}

func TestDeferOrder(t *testing.T) {
	got := DeferOrder()
	want := []string{"third", "second", "first"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DeferOrder() = %v, want %v", got, want)
	}
}

func TestSafeCall(t *testing.T) {
	if SafeCall(func() {}) {
		t.Fatal("SafeCall() = true, want false")
	}

	if !SafeCall(func() { panic("boom") }) {
		t.Fatal("SafeCall() = false, want true")
	}
}
