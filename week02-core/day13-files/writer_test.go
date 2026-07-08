package files

import (
	"os"
	"testing"
)

func TestWriteAll(t *testing.T) {
	// Create a temp file path
	file, err := os.CreateTemp("", "test*.txt")
	if err != nil {
		t.Fatalf("create temp file error = %v", err)
	}
	path := file.Name()
	file.Close()
	defer os.Remove(path)

	// Test WriteAll
	content := "hello world"
	err = WriteAll(path, content)
	if err != nil {
		t.Fatalf("WriteAll error = %v", err)
	}

	// Verify content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file error = %v", err)
	}

	if string(data) != content {
		t.Fatalf("file content = %q, want %q", string(data), content)
	}
}

func TestWriteLines(t *testing.T) {
	// Create a temp file path
	file, err := os.CreateTemp("", "test*.txt")
	if err != nil {
		t.Fatalf("create temp file error = %v", err)
	}
	path := file.Name()
	file.Close()
	defer os.Remove(path)

	// Test WriteLines
	lines := []string{"line1", "line2", "line3"}
	err = WriteLines(path, lines)
	if err != nil {
		t.Fatalf("WriteLines error = %v", err)
	}

	// Verify content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file error = %v", err)
	}

	want := "line1\nline2\nline3\n"
	if string(data) != want {
		t.Fatalf("file content = %q, want %q", string(data), want)
	}
}

func TestWriteLinesEmpty(t *testing.T) {
	// Create a temp file path
	file, err := os.CreateTemp("", "test*.txt")
	if err != nil {
		t.Fatalf("create temp file error = %v", err)
	}
	path := file.Name()
	file.Close()
	defer os.Remove(path)

	// Test WriteLines with empty slice
	err = WriteLines(path, []string{})
	if err != nil {
		t.Fatalf("WriteLines error = %v", err)
	}

	// Verify content is empty
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file error = %v", err)
	}

	if string(data) != "" {
		t.Fatalf("file content = %q, want empty", string(data))
	}
}
