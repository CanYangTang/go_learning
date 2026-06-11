package goroutine

import (
	"sync"
	"time"
)

type Task struct {
	Name     string
	Duration time.Duration
}

func RunSerial(tasks []Task) []string {
	var results []string
	for _, task := range tasks {
		time.Sleep(task.Duration)
		results = append(results, task.Name)
	}
	return results
}

func RunConcurrent(tasks []Task) []string {
	var wg sync.WaitGroup
	results := make([]string, len(tasks))
	for i, task := range tasks {
		wg.Add(1)
		go func(i int, task Task) {
			defer wg.Done()
			time.Sleep(task.Duration)
			results[i] = task.Name
		}(i, task)
	}
	wg.Wait()
	return results
}

func MeasureSerial(tasks []Task) time.Duration {
	start := time.Now()
	RunSerial(tasks)
	return time.Since(start)
}

func MeasureConcurrent(tasks []Task) time.Duration {
	start := time.Now()
	RunConcurrent(tasks)
	return time.Since(start)
}
