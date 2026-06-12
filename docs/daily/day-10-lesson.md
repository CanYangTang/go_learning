# Day 10 教案：并发任务队列

## 学习目标

学完今天，需要能够做到：

1. 理解 producer-consumer 模型。
2. 使用 jobs channel 分发任务。
3. 使用 results channel 收集结果。
4. 使用多个 worker 并发处理任务。
5. 使用 `close(jobs)` 通知 worker 没有更多任务。
6. 使用 `sync.WaitGroup` 等待所有 worker 退出。
7. 理解 graceful shutdown 的基本思路。
8. 对比单 worker 和多 worker 的处理耗时。

---

## Day 10 的位置

Day 8 学了 goroutine 和 `sync.WaitGroup`。

Day 9 学了 channel、`close`、`for range`、`select` 和 `time.After`。

Day 10 会把它们组合起来，写一个小型并发任务队列。

这也是 Go 并发最常见的组合方式之一：

```text
producer -> jobs channel -> workers -> results channel -> collector
```

---

## 什么是 producer-consumer 模型

producer-consumer 是生产者-消费者模型。

- producer：生产任务，把任务放进队列。
- consumer：消费任务，从队列取任务并处理。

在 Go 中，channel 很适合表示这个队列。

```text
producer 发送任务到 jobs channel
worker 从 jobs channel 接收任务
worker 处理后把结果发送到 results channel
```

---

## 为什么需要任务队列

如果有很多任务，例如：

- 处理多个文件
- 抓取多个 URL
- 发送多条消息
- 执行多个耗时计算

串行执行会很慢。

但如果每个任务都直接启动一个 goroutine，也可能太多，难以控制。

任务队列的好处是：

- 可以限制 worker 数量。
- 可以控制并发度。
- 可以统一收集结果。
- 可以在任务发送完后优雅退出 worker。

---

## jobs channel

jobs channel 用来分发任务。

例如：

```go
jobs := make(chan Job)
```

producer 可以发送任务：

```go
jobs <- job
```

worker 可以接收任务：

```go
for job := range jobs {
    // process job
}
```

当 producer 发送完所有任务后，应该关闭 jobs：

```go
close(jobs)
```

这样 worker 的 `for range jobs` 才能自然结束。

---

## results channel

results channel 用来收集处理结果。

例如：

```go
results := make(chan Result)
```

worker 处理完任务后发送结果：

```go
results <- result
```

collector 从 results 中接收：

```go
result := <-results
```

今天为了测试稳定，可以按任务数量接收固定次数结果，而不是必须关闭 results。

---

## worker 是什么

worker 是持续从 jobs channel 取任务并处理的 goroutine。

典型结构：

```go
func worker(jobs <-chan Job, results chan<- Result) {
    for job := range jobs {
        result := process(job)
        results <- result
    }
}
```

这里用了 channel 方向：

- `<-chan Job`：只能接收 Job。
- `chan<- Result`：只能发送 Result。

方向不是必须写，但写了能让函数意图更清楚。

---

## channel 方向

普通 channel：

```go
chan Job
```

可以发送，也可以接收。

只接收：

```go
<-chan Job
```

只发送：

```go
chan<- Result
```

示例：

```go
func worker(jobs <-chan Job, results chan<- Result) {
    for job := range jobs {
        results <- Result{ID: job.ID}
    }
}
```

这样 worker 不能误向 jobs 发送数据，也不能误从 results 接收数据。

---

## close(jobs) 的作用

worker 常见写法：

```go
for job := range jobs {
    // process job
}
```

这个循环什么时候结束？

答案是：当 jobs 被关闭，并且已发送的数据都读完。

所以 producer 发送完所有任务后要：

```go
close(jobs)
```

否则 worker 会一直等下一个任务。

---

## WaitGroup 在任务队列中的作用

`WaitGroup` 用来等待所有 worker 退出。

```go
var wg sync.WaitGroup

for i := 0; i < workerCount; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        worker(jobs, results)
    }()
}

close(jobs)
wg.Wait()
```

常见用途：

- 等 worker 把 jobs 里的任务处理完。
- 等所有 worker 从 `for range jobs` 中退出。
- 在所有 worker 退出后再关闭 results。

---

## 为什么有时要 close(results)

如果 collector 使用：

```go
for result := range results {
    // collect
}
```

那么就必须在所有 worker 不再发送结果后关闭 results。

常见写法：

```go
go func() {
    wg.Wait()
    close(results)
}()
```

意思是：等所有 worker 退出后，说明不会再发送结果，可以关闭 results。

今天为了降低复杂度，也可以按任务数量接收固定次数结果：

```go
for i := 0; i < len(jobs); i++ {
    result := <-results
}
```

这种方式可以不关闭 results。

---

## graceful shutdown 是什么

graceful shutdown 可以理解为“优雅停止”。

不是突然打断 worker，而是：

1. 不再发送新任务。
2. 关闭 jobs channel。
3. worker 处理完已经收到的任务。
4. worker 从 `for range jobs` 中退出。
5. 主流程等待所有 worker 退出。
6. 收集完结果后返回。

今天的 `close(jobs)` + `WaitGroup` 就是一个最小版本的 graceful shutdown。

---

## worker 数量和并发度

worker 数量决定并发度。

```go
workerCount := 3
```

表示最多 3 个任务同时被处理。

如果任务很多：

- 1 个 worker：接近串行。
- 多个 worker：可以并发处理。
- worker 太多：可能浪费资源或造成压力。

实际项目里 worker 数量通常需要根据任务类型和系统资源调整。

---

## 任务和结果设计

今天建议定义：

```go
type Job struct {
    ID       int
    Duration time.Duration
}

type Result struct {
    JobID int
    Value int
}
```

处理规则：

```text
处理 Job 时 sleep Duration，然后返回 Result{JobID: job.ID, Value: job.ID * 2}
```

这样测试容易验证。

---

## 今日代码练习设计

今天创建：

```text
week02-core/day10-task-queue/
├── queue.go
└── queue_test.go
```

实现：

```go
type Job struct {
    ID       int
    Duration time.Duration
}

type Result struct {
    JobID int
    Value int
}

func ProcessJobs(jobs []Job, workerCount int) []Result
func MeasureProcessJobs(jobs []Job, workerCount int) time.Duration
```

---

## ProcessJobs

目标：用固定数量 worker 并发处理任务。

规则：

1. 如果 `workerCount <= 0`，当作 1 个 worker。
2. 创建 jobs channel。
3. 创建 results channel。
4. 启动 `workerCount` 个 worker。
5. producer 把所有 job 发送到 jobs channel。
6. 发送完后关闭 jobs channel。
7. 收集和 jobs 数量相同的结果。
8. 等 worker 全部退出。
9. 返回结果。

注意：多个 worker 并发处理，结果顺序不一定等于输入顺序。

测试时可以按 `JobID` 排序后比较。

---

## worker 函数参考思路

可以写一个未导出的 helper：

```go
func worker(jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
    defer wg.Done()
    for job := range jobs {
        time.Sleep(job.Duration)
        results <- Result{JobID: job.ID, Value: job.ID * 2}
    }
}
```

这里：

- `jobs <-chan Job` 只能接收任务。
- `results chan<- Result` 只能发送结果。
- `defer wg.Done()` 确保 worker 退出时通知 WaitGroup。

---

## 收集结果

因为知道任务数量，所以可以固定接收：

```go
collected := make([]Result, 0, len(jobs))
for i := 0; i < len(jobs); i++ {
    collected = append(collected, <-results)
}
```

这样不依赖关闭 results。

收集完后：

```go
wg.Wait()
```

确保所有 worker 已经从 jobs channel 退出。

---

## MeasureProcessJobs

和 Day 8 类似：

```go
func MeasureProcessJobs(jobs []Job, workerCount int) time.Duration {
    start := time.Now()
    ProcessJobs(jobs, workerCount)
    return time.Since(start)
}
```

用于比较不同 worker 数量的耗时。

---

## 今日验收标准

完成后应该满足：

1. 创建 `week02-core/day10-task-queue/`。
2. 实现 `Job`、`Result`、`ProcessJobs`、`MeasureProcessJobs`。
3. 至少写 5 组测试。
4. `workerCount <= 0` 时能退化为 1 个 worker。
5. 能正确处理空任务。
6. 多 worker 处理结果正确。
7. 多 worker 耗时明显少于单 worker。
8. `go test ./...` 通过。
9. 能解释 producer-consumer、jobs channel、results channel、worker、graceful shutdown。

---

## 今天最容易踩的坑

### 坑 1：忘记 close(jobs)

worker 使用 `for job := range jobs`，如果 jobs 不关闭，worker 会一直等。

---

### 坑 2：结果顺序不稳定

多个 worker 并发处理，结果完成顺序不一定等于输入顺序。

测试时不要直接依赖顺序，可以按 `JobID` 排序后比较。

---

### 坑 3：workerCount 为 0

如果 worker 数量是 0，没有人消费 jobs，发送任务会阻塞。

今天要求 `workerCount <= 0` 时当作 1。

---

### 坑 4：results 没人接收导致 worker 阻塞

如果 worker 发送 results，但主流程没有及时接收，可能阻塞。

今天可以用容量为 `len(jobs)` 的 buffered results channel，降低阻塞风险。

---

### 坑 5：过早关闭 results

如果 worker 还在发送结果，主流程就关闭 results，会 panic。

如果要关闭 results，必须确保所有 worker 都已经退出。
