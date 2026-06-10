package packages

import (
	"fmt"
	"github.com/CanYangTang/go_learning/pkg/mathutil"
	"strings"
)

var status string

func init() {
	status = "ready"
}

func NormalizeName(name string) string {
	return strings.TrimSpace(name)
}

func UserLabel(name string, age int) string {
	return fmt.Sprintf("%s(%d)", NormalizeName(name), age)
}

func DoubleAge(age int) int {
	return mathutil.Double(age)
}

func IsAdultAgeEven(age int) bool {
	return age >= 18 && mathutil.IsEven(age)
}

func PackageStatus() string {
	return status
}
