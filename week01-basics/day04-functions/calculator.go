package functions

func Add(a, b float64) float64 {
	return a + b
}

func Subtract(a, b float64) float64 {
	return a - b
}

func Multiply(a, b float64) float64 {
	return a * b
}

func Divide(a, b float64) (float64, bool) {
	if b == 0 {
		return 0, false
	}
	return a / b, true
}

func Sum(nums ...float64) float64 {
	sum := 0.0
	for _, num := range nums {
		sum += num
	}
	return sum
}

func Apply(a, b float64, op func(float64, float64) float64) float64 {
	return op(a, b)
}

func NewCounter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}
