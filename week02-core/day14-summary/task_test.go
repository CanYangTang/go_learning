package summary

import (
	"testing"
)

func TestProcessTask(t *testing.T) {
	task := Task{ID: 1, Data: "hello"}
	result := ProcessTask(task)

	if result.ID != 1 {
		t.Fatalf("result.ID = %d, want 1", result.ID)
	}

	expectedOutput := "processed hello"
	if result.Output != expectedOutput {
		t.Fatalf("result.Output = %q, want %q", result.Output, expectedOutput)
	}
}

func TestProcessBatchSuccess(t *testing.T) {
	tasks := []Task{
		{ID: 1, Data: "a"},
		{ID: 2, Data: "b"},
		{ID: 3, Data: "c"},
	}

	results := ProcessBatch(tasks, 2)

	if len(results) != 3 {
		t.Fatalf("len(results) = %d, want 3", len(results))
	}

	// Results should be sorted by ID
	for i, result := range results {
		if result.ID != i+1 {
			t.Fatalf("result.ID = %d, want %d", result.ID, i+1)
		}
		expectedOutput := "processed " + tasks[i].Data
		if result.Output != expectedOutput {
			t.Fatalf("result.Output = %q, want %q", result.Output, expectedOutput)
		}
	}
}

func TestProcessBatchEmpty(t *testing.T) {
	results := ProcessBatch([]Task{}, 3)

	if len(results) != 0 {
		t.Fatalf("len(results) = %d, want 0", len(results))
	}
}

func TestProcessBatchFallbackWorkerCount(t *testing.T) {
	tasks := []Task{
		{ID: 1, Data: "x"},
	}

	// workerCount 0 should fallback to 1
	results := ProcessBatch(tasks, 0)

	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}

	if results[0].ID != 1 {
		t.Fatalf("result.ID = %d, want 1", results[0].ID)
	}
}
