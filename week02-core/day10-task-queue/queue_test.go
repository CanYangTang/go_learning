package taskqueue

import (
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestProcessJobs(t *testing.T) {
	jobs := []Job{
		{ID: 1, Duration: time.Millisecond},
		{ID: 2, Duration: time.Millisecond},
		{ID: 3, Duration: time.Millisecond},
	}

	got := ProcessJobs(jobs, 2)
	want := []Result{
		{JobID: 1, Value: 2},
		{JobID: 2, Value: 4},
		{JobID: 3, Value: 6},
	}

	assertResultsEqual(t, got, want)
}

func TestProcessJobsEmpty(t *testing.T) {
	got := ProcessJobs(nil, 3)

	if len(got) != 0 {
		t.Fatalf("ProcessJobs(nil) length = %v, want 0", len(got))
	}
}

func TestProcessJobsFallbackWorkerCount(t *testing.T) {
	jobs := []Job{
		{ID: 1, Duration: time.Millisecond},
		{ID: 2, Duration: time.Millisecond},
	}

	got := ProcessJobs(jobs, 0)
	want := []Result{
		{JobID: 1, Value: 2},
		{JobID: 2, Value: 4},
	}

	assertResultsEqual(t, got, want)
}

func TestMeasureProcessJobs(t *testing.T) {
	jobs := []Job{
		{ID: 1, Duration: 30 * time.Millisecond},
		{ID: 2, Duration: 30 * time.Millisecond},
		{ID: 3, Duration: 30 * time.Millisecond},
	}

	singleWorker := MeasureProcessJobs(jobs, 1)
	multiWorker := MeasureProcessJobs(jobs, 3)

	if singleWorker <= 0 {
		t.Fatalf("MeasureProcessJobs(jobs, 1) = %v, want positive duration", singleWorker)
	}

	if multiWorker <= 0 {
		t.Fatalf("MeasureProcessJobs(jobs, 3) = %v, want positive duration", multiWorker)
	}

	if multiWorker >= singleWorker/2 {
		t.Fatalf("MeasureProcessJobs(jobs, 3) = %v, want less than half of single worker duration %v", multiWorker, singleWorker)
	}
}

func TestProcessJobsManyWorkers(t *testing.T) {
	jobs := []Job{
		{ID: 1, Duration: time.Millisecond},
		{ID: 2, Duration: time.Millisecond},
	}

	got := ProcessJobs(jobs, 10)
	want := []Result{
		{JobID: 1, Value: 2},
		{JobID: 2, Value: 4},
	}

	assertResultsEqual(t, got, want)
}

func assertResultsEqual(t *testing.T, got, want []Result) {
	t.Helper()

	sortResults(got)
	sortResults(want)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("results = %v, want %v", got, want)
	}
}

func sortResults(results []Result) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].JobID < results[j].JobID
	})
}
