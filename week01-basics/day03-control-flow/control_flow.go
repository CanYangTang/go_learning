package controlflow

import (
	"fmt"
	"strings"
)

func Grade(score int) string {
	if score < 0 || score > 100 {
		return "Invalid"
	} else if score >= 90 {
		return "A"
	} else if score >= 80 {
		return "B"
	} else if score >= 60 {
		return "C"
	} else {
		return "D"
	}
}

func DayType(day string) string {
	switch day {
	case "Mon", "Tue", "Wed", "Thu", "Fri":
		return "weekday"
	case "Sat", "Sun":
		return "weekend"
	default:
		return "unknown"
	}
}

func EvenNumbers(n int) []int {
	var res []int
	for i := 1; i <= n; i++ {
		if i%2 != 0 {
			continue
		}
		res = append(res, i)
	}
	return res
}

func SumTo(n int) int {
	sum := 0
	for i := 0; i <= n; i++ {
		sum += i
	}
	return sum
}

func MultiplicationTable() []string {
	var rows []string
	for i := 1; i <= 9; i++ {
		var cols []string
		for j := 1; j <= i; j++ {
			cols = append(cols, fmt.Sprintf("%d*%d=%d", j, i, i*j))
		}
		rows = append(rows, strings.Join(cols, " "))
	}
	return rows
}
