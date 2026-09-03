package config

import (
	"strings"
	"testing"
)

func TestDefaultDSN(t *testing.T) {
	dsn := DefaultDSN()

	fragments := []string{
		"root:",
		"@tcp(127.0.0.1:3306)/",
		"go_learning",
		"charset=utf8mb4",
		"parseTime=True",
	}
	for _, fragment := range fragments {
		if !strings.Contains(dsn, fragment) {
			t.Fatalf("DSN %q does not contain %q", dsn, fragment)
		}
	}
}
