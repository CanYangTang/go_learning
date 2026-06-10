package errors

import (
	stderrors "errors"
	"fmt"
)

var ErrDivideByZero = stderrors.New("divide by zero")
var ErrUnsupportedOperation = stderrors.New("invalid operator")

func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivideByZero
	}
	return a / b, nil
}

func Calculate(a, b float64, op string) (float64, error) {
	switch op {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/":
		result, err := Divide(a, b)
		if err != nil {
			return 0, fmt.Errorf("calculate failed: %w", err)
		}
		return result, nil
	default:
		return 0, fmt.Errorf("calculate failed: %w", ErrUnsupportedOperation)
	}
}

func IsDivideByZero(err error) bool {
	return stderrors.Is(err, ErrDivideByZero)
}

func DeferOrder() (order []string) {
	defer func() {
		order = append(order, "first")
	}()
	defer func() { order = append(order, "second") }()
	defer func() { order = append(order, "third") }()
	return order
}

func SafeCall(fn func()) (panicked bool) {
	defer func() {
		if r := recover(); r != nil {
			panicked = true
		}
	}()
	fn()
	return false
}
