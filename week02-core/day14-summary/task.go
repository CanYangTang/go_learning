package summary

import (
	"sort"
	"sync"
	"time"
)

type Task struct {
	ID   int    `json:"id"`
	Data string `json:"data"`
}

type Result struct {
	ID     int    `json:"id"`
	Output string `json:"output"`
}

type BatchRequest struct {
	Tasks []Task `json:"tasks"`
}

type BatchResponse struct {
	Results []Result `json:"results"`
}

// ProcessTask processes a single task and returns the result.
func ProcessTask(task Task) Result {
	// TODO: implement
	// Simulate processing with a small delay
	// Return Result with ID and "processed " + task.Data
	time.Sleep(5 * time.Millisecond)
	return Result{
		ID:     task.ID,
		Output: "processed " + task.Data,
	}
}

// ProcessBatch processes multiple tasks concurrently with a worker pool.
func ProcessBatch(tasks []Task, workerCount int) []Result {
	// TODO: implement
	// 1. If workerCount <= 0, set to 1
	// 2. Create jobs channel and results channel
	// 3. Start workerCount workers, each reads from jobs and sends to results
	// 4. Send all tasks to jobs channel
	// 5. Close jobs channel
	// 6. Wait for all workers to finish, then close results channel
	// 7. Collect results from results channel
	// 8. Sort results by ID
	// 9. Return sorted results
	if workerCount <= 0 {
		workerCount = 1
	}
	jobs := make(chan Task, len(tasks))
	results := make(chan Result, len(tasks))

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				results <- ProcessTask(job)
			}
		}()
	}
	for _, task := range tasks {
		jobs <- task
	}
	close(jobs)
	go func() {
		wg.Wait()
		close(results)
	}()

	var outputs []Result
	for result := range results {
		outputs = append(outputs, result)
	}

	sort.Slice(outputs, func(x, y int) bool {
		return outputs[x].ID < outputs[y].ID
	})

	return outputs
}
