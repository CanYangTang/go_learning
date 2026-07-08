package files

import (
	"os"
	"testing"
)

func TestReadAll(t *testing.T) {
	// Create a temp file with content
	file, err := os.CreateTemp("", "test*.txt")
	if err != nil {
		t.Fatalf("create temp file error = %v", err)
	}
	defer os.Remove(file.Name())

	content := "hello world\nsecond line"
	file.WriteString(content)
	file.Close()

	// Test ReadAll
	got, err := ReadAll(file.Name())
	if err != nil {
		t.Fatalf("ReadAll error = %v", err)
	}

	if got != content {
		t.Fatalf("ReadAll = %q, want %q", got, content)
	}
}

func TestReadAllFileNotExist(t *testing.T) {
	_, err := ReadAll("/nonexistent/file.txt")
	if err == nil {
		t.Fatalf("ReadAll should return error for nonexistent file")
	}
}

func TestReadLines(t *testing.T) {
	// Create a temp file with multiple lines
	file, err := os.CreateTemp("", "test*.txt")
	if err != nil {
		t.Fatalf("create temp file error = %v", err)
	}
	defer os.Remove(file.Name())

	file.WriteString("line1\nline2\nline3\n")
	file.Close()

	// Test ReadLines
	got, err := ReadLines(file.Name())
	if err != nil {
		t.Fatalf("ReadLines error = %v", err)
	}

	want := []string{"line1", "line2", "line3"}
	if len(got) != len(want) {
		t.Fatalf("ReadLines length = %d, want %d", len(got), len(want))
	}

	for i, line := range got {
		if line != want[i] {
			t.Fatalf("ReadLines[%d] = %q, want %q", i, line, want[i])
		}
	}
}

func TestReadLinesEmptyFile(t *testing.T) {
	// Create an empty temp file
	file, err := os.CreateTemp("", "test*.txt")
	if err != nil {
		t.Fatalf("create temp file error = %v", err)
	}
	defer os.Remove(file.Name())
	file.Close()

	// Test ReadLines
	got, err := ReadLines(file.Name())
	if err != nil {
		t.Fatalf("ReadLines error = %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("ReadLines length = %d, want 0", len(got))
	}
}
