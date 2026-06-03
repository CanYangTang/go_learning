package syntax

import (
	"fmt"
	"strconv"
)

func UserProfile(name string, age int) string {
	return fmt.Sprintf("%s is %d years old", name, age)
}

func RectangleArea(width, height float64) float64 {
	return width * height
}

func Average(total int, count int) float64 {
	return float64(total) / float64(count)
}

func IsAdult(age int) bool {
	return age >= 18
}

func FormatScore(score int) string {
	return "score=" + strconv.Itoa(score)
}
