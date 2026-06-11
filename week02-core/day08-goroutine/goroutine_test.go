package goroutine

import (
	"reflect"
	"testing"
	"time"
)

func TestRunSerial(t *testing.T) {
	tasks := []Task{
		{Name: "first", Duration: time.Millisecond},
		{Name: "second", Duration: time.Millisecond},
		{Name: "third", Duration: time.Millisecond},
	}

	got := RunSerial(tasks)
	want := []string{"first", "second", "third"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RunSerial() = %v, want %v", got, want)
	}
}

func TestRunSerialEmpty(t *testing.T) {
	got := RunSerial(nil)

	if len(got) != 0 {
		t.Fatalf("RunSerial(nil) length = %v, want 0", len(got))
	}
}

func TestRunConcurrent(t *testing.T) {
	tasks := []Task{
		{Name: "first", Duration: 3 * time.Millisecond},
		{Name: "second", Duration: time.Millisecond},
		{Name: "third", Duration: 2 * time.Millisecond},
	}

	got := RunConcurrent(tasks)
	want := []string{"first", "second", "third"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RunConcurrent() = %v, want %v", got, want)
	}
}

func TestRunConcurrentEmpty(t *testing.T) {
	got := RunConcurrent(nil)

	if len(got) != 0 {
		t.Fatalf("RunConcurrent(nil) length = %v, want 0", len(got))
	}
}

func TestMeasureConcurrentFasterThanSerial(t *testing.T) {
	tasks := []Task{
		{Name: "first", Duration: 30 * time.Millisecond},
		{Name: "second", Duration: 30 * time.Millisecond},
		{Name: "third", Duration: 30 * time.Millisecond},
	}

	serial := MeasureSerial(tasks)
	concurrent := MeasureConcurrent(tasks)

	if serial <= 0 {
		t.Fatalf("MeasureSerial() = %v, want positive duration", serial)
	}

	if concurrent <= 0 {
		t.Fatalf("MeasureConcurrent() = %v, want positive duration", concurrent)
	}

	if concurrent >= serial/2 {
		t.Fatalf("MeasureConcurrent() = %v, want less than half of serial duration %v", concurrent, serial)
	}
}
