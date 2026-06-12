package taskqueue

import (
	"sync"
	"time"
)

type Job struct {
	ID       int
	Duration time.Duration
}

type Result struct {
	JobID int
	Value int
}

func ProcessJobs(jobs []Job, workerCount int) []Result {
	if workerCount <= 0 {
		workerCount = 1
	}
	jobsChan := make(chan Job)
	resultsChan := make(chan Result, len(jobs))
	wg := sync.WaitGroup{}
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobsChan {
				time.Sleep(job.Duration)
				resultsChan <- Result{JobID: job.ID, Value: job.ID * 2}
			}
		}()
	}
	for _, job := range jobs {
		jobsChan <- job
	}
	close(jobsChan)
	results := []Result{}
	go func() {
		wg.Wait()
		close(resultsChan)
	}()
	for result := range resultsChan {
		results = append(results, result)
	}
	return results
}

func MeasureProcessJobs(jobs []Job, workerCount int) time.Duration {
	start := time.Now()
	ProcessJobs(jobs, workerCount)
	return time.Since(start)
}
