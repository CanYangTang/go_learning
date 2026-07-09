# Day 14 教案：HTTP + 并发综合练习

## 学习目标

学完今天，需要能够做到：

1. 把 HTTP server、JSON 处理、并发任务队列组合起来。
2. 实现一个并发 JSON 处理的 HTTP handler。
3. 使用 `sync.WaitGroup` 和 channel 协调多个 worker。
4. 编写综合练习的测试。
5. 完成 Week 2 的知识串联和巩固。

---

## Day 14 的位置

Day 14 是 Week 2 的综合练习日，会把本周学过的内容串联起来：

- Day 8：goroutine 和 `sync.WaitGroup`
- Day 9：channel 通信
- Day 10：并发任务队列（producer-consumer）
- Day 11：HTTP server 和 handler
- Day 12：JSON 请求/响应处理
- Day 13：文件操作

今天的目标是实现一个能并发处理多个 JSON 请求任务的 HTTP handler。

---

## 综合练习设计

设想一个场景：

客户端发送一个 JSON 请求，包含多个任务：

```json
{
  "tasks": [
    {"id": 1, "data": "task1"},
    {"id": 2, "data": "task2"},
    {"id": 3, "data": "task3"}
  ]
}
```

服务端并发处理这些任务，返回处理结果：

```json
{
  "results": [
    {"id": 1, "output": "processed task1"},
    {"id": 2, "output": "processed task2"},
    {"id": 3, "output": "processed task3"}
  ]
}
```

---

## 综合练习的结构

建议创建：

```text
week02-core/day14-summary/
├── handler.go
├── handler_test.go
├── task.go
└── task_test.go
```

---

## Task 结构

定义任务和结果结构：

```go
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
```

---

## ProcessTask 函数

单个任务的处理函数：

```go
func ProcessTask(task Task) Result {
    // 模拟处理，比如加一点延迟
    time.Sleep(10 * time.Millisecond)
    return Result{
        ID:     task.ID,
        Output: "processed " + task.Data,
    }
}
```

---

## ProcessBatch 并发处理

批量并发处理任务：

```go
func ProcessBatch(tasks []Task, workerCount int) []Result {
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
            for task := range jobs {
                results <- ProcessTask(task)
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

    // 按 ID 排序，确保结果顺序稳定
    sort.Slice(outputs, func(i, j int) bool {
        return outputs[i].ID < outputs[j].ID
    })

    return outputs
}
```

这是 Day 10 任务队列模式的应用。

---

## BatchHandler

HTTP handler：

```go
func BatchHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    defer r.Body.Close()

    var req BatchRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "invalid json"})
        return
    }

    results := ProcessBatch(req.Tasks, 3)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(BatchResponse{Results: results})
}
```

---

## 测试设计

至少写 3 组测试：

1. `TestProcessBatchSuccess` — 多任务并发处理成功。
2. `TestProcessBatchEmpty` — 空任务列表返回空结果。
3. `TestBatchHandlerSuccess` — HTTP handler 成功处理请求。

示例：

```go
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

    // 按 ID 检查结果
    for i, result := range results {
        if result.ID != i+1 {
            t.Fatalf("result.ID = %d, want %d", result.ID, i+1)
        }
    }
}

func TestBatchHandlerSuccess(t *testing.T) {
    body := `{"tasks":[{"id":1,"data":"x"},{"id":2,"data":"y"}]}`
    req := httptest.NewRequest(http.MethodPost, "/batch", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    recorder := httptest.NewRecorder()

    BatchHandler(recorder, req)

    if recorder.Code != http.StatusOK {
        t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
    }

    var resp BatchResponse
    if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
        t.Fatalf("decode error = %v", err)
    }

    if len(resp.Results) != 2 {
        t.Fatalf("len(results) = %d, want 2", len(resp.Results))
    }
}
```

---

## 今日验收标准

完成后应该满足：

1. 创建 `week02-core/day14-summary/`。
2. 实现 `Task`、`Result`、`BatchRequest`、`BatchResponse` 结构体。
3. 实现 `ProcessTask`、`ProcessBatch`、`BatchHandler`。
4. 至少写 3 组测试。
5. `go test ./week02-core/day14-summary` 通过。
6. `make test` 通过。
7. 能解释任务队列、JSON handler、并发协调如何组合。

---

## 可选挑战题：context timeout

Issue 的可选挑战是增加 context timeout 或 cancellation。

可以在 `ProcessBatch` 中加入 context 参数：

```go
func ProcessBatch(ctx context.Context, tasks []Task, workerCount int) []Result
```

Worker 检查 `ctx.Done()`：

```go
for task := range jobs {
    select {
    case <-ctx.Done():
        return // 提前退出
    default:
        results <- ProcessTask(task)
    }
}
```

今天可以先完成基础版本，挑战题作为可选扩展。

---

## Week 2 知识串联总结

| Day | 主题 | 核心知识点 |
|-----|------|-----------|
| 8 | goroutine | `go fn()`, `sync.WaitGroup`, `Add/Done/Wait` |
| 9 | channel | 无缓冲/有缓冲 channel, `close`, `for range`, `select` |
| 10 | 任务队列 | producer-consumer, jobs/results channel, worker pool |
| 11 | HTTP server | `net/http`, `ServeMux`, handler, `ResponseWriter/Request` |
| 12 | JSON | `Marshal/Unmarshal`, struct tag, `Encoder/Decoder` |
| 13 | 文件操作 | `os.ReadFile/WriteFile`, `bufio.Scanner/Writer` |
| 14 | 综合 | HTTP + JSON + 并发组合 |

今天完成后，Week 2 的基础能力就串联起来了。

---

## 今天最容易踩的坑

### 坑 1：忘记排序结果

并发处理的结果顺序不稳定，必须按 `ID` 排序后再返回或比较。

---

### 坑 2：`defer r.Body.Close()` 放错位置

`defer r.Body.Close()` 应该放在方法检查之后、解码之前。如果放在解码之后，JSON 解析失败时提前 return，`defer` 未登记，`r.Body` 不会关闭，导致资源泄漏。

---

### 坑 3：`workerCount <= 0` 没有兜底

如果没有 worker，jobs channel 没人消费，发送会阻塞。

---

### 坑 4：忘记 `close(jobs)` 和 `close(results)`

不关闭 channel，worker 和 collector 会一直等待。

---

### 坑 5：测试中忘记设置 `Content-Type`

测试 HTTP handler 时，JSON 请求需要设置：

```go
req.Header.Set("Content-Type", "application/json")
```