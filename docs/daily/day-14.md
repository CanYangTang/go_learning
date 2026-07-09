# Day 14 学习记录

## 日期

2026-07-08

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/14

## 今日教案

- 教案文档：`docs/daily/day-14-lesson.md`

## 核心任务

- 把 HTTP server、JSON 处理、并发任务队列组合起来。
- 实现一个并发 JSON 处理的 HTTP handler。
- 使用 `sync.WaitGroup` 和 channel 协调多个 worker。
- 完成 Week 2 的知识串联和巩固。

## 验收标准

- 创建 `week02-core/day14-summary/`。
- 实现 `Task`、`Result`、`BatchRequest`、`BatchResponse` 结构体。
- 实现 `ProcessTask`、`ProcessBatch`、`BatchHandler`。
- 至少写 3 组测试。
- `go test ./week02-core/day14-summary` 通过。
- `make test` 通过。
- 能解释任务队列、JSON handler、并发协调如何组合。

## 可选挑战题

- 增加 context timeout 或 cancellation。

## 答疑记录

- 待记录。

## 今日产出

- 创建 `week02-core/day14-summary/` 目录。
- 实现 `Task`、`Result`、`BatchRequest`、`BatchResponse` 结构体，带 `json` tag。
- 实现 `ProcessTask`：模拟处理，返回 `"processed " + task.Data`。
- 实现 `ProcessBatch`：使用 worker pool 并发处理任务，结果按 `ID` 排序。
- 实现 `BatchHandler`：HTTP handler，解析 JSON 请求，调用 ProcessBatch，返回结果。
- 编写测试覆盖单任务处理、批量并发处理、空任务、workerCount 兜底、HTTP 成功/错误/方法检查。

## 运行过的命令

```bash
go test ./week02-core/day14-summary
make fmt
make test
make vet
```

## 代码 Review 结论

- `ProcessTask` 用 `time.Sleep(5ms)` 模拟处理延迟，返回 `Result{ID, "processed " + Data}`。
- `ProcessBatch` 使用 Day 10 的 producer-consumer 模式：jobs channel 分发任务，worker pool 并发处理，results channel 收集结果。
- `workerCount <= 0` 时退化为 1 个 worker，避免无人消费导致阻塞。
- `close(jobs)` 让 worker 的 `for range jobs` 自然退出。
- `go func() { wg.Wait(); close(results) }()` 在所有 worker 退出后关闭 results channel。
- `sort.Slice` 按 `ID` 排序结果，确保测试顺序稳定。
- `BatchHandler` 检查 POST 方法，解析 JSON，设置 `Content-Type`，返回 200 和 BatchResponse。
- `go test ./week02-core/day14-summary`、`make fmt`、`make test`、`make vet` 均已通过。

## 今日小测试

1. `ProcessBatch` 中 jobs channel 和 results channel 分别有什么作用？
   - 回答：jobs channel：主 goroutine 往里发任务，worker goroutine 从里取任务。起到任务分发队列的作用。results channel：worker goroutine 把处理结果往里发，主 goroutine 最后统一收集。两个 channel 把"分发任务"和"收集结果"解耦。
   - 结果：通过。
   - 标准答案：jobs channel 是任务分发队列，producer 发送任务，worker 消费任务；results channel 是结果收集队列，worker 发送结果，主 goroutine 收集结果。两者解耦了任务分发和结果收集。
2. 为什么需要 `go func() { wg.Wait(); close(results) }()` 而不是直接 `wg.Wait()` 后再 `close(results)`？
   - 回答：主 goroutine 负责收集，另一个 goroutine 负责等待并关闭，两者并发进行，不会死锁。
   - 结果：通过。
   - 标准答案：主 goroutine 在 `for result := range results` 中接收结果，如果先 `wg.Wait()` 再 `close(results)`，主 goroutine 会阻塞在 `range results`，而 worker 发送完结果后也没人关闭 results，导致死锁。用单独 goroutine 等待并关闭，让主 goroutine 可以并发收集。
3. `sort.Slice(outputs, func(x, y int) bool {...})` 做了什么？为什么并发处理的结果需要排序？
   - 回答：按 ID 升序排列结果切片。为什么需要排序：worker 是并发执行的，哪个先完成完全取决于调度和任务耗时，结果进入 results 的顺序是不确定的。
   - 结果：通过。
   - 标准答案：`sort.Slice` 按 ID 升序排序结果。并发处理时，worker 完成顺序不确定，results channel 收到的顺序也不确定，所以需要排序后才能与期望结果比较。
4. `BatchHandler` 中 `defer r.Body.Close()` 放在哪里？为什么？
   - 回答：放在函数最开头，获取到 `r.Body` 之后立刻登记。原因：如果放在解码之后，解码失败时走 return，defer 还没登记，`r.Body` 就永远不会关闭，导致资源泄漏。
   - 结果：回答正确，但实现有 bug。
   - 标准答案：`defer r.Body.Close()` 应该放在方法检查之后、解码之前。当前实现放在解码之后（第 29 行），如果解码失败在第 27 行 return，defer 未登记，`r.Body` 不会关闭。正确位置应该是第 21 行（方法检查之后）。
5. 这个综合练习串联了 Week 2 哪几天的知识点？
   - 回答：goroutine、WaitGroup、channel。HTTP handler、请求/响应。json.Decode / json.Encode。worker pool 模式。sort.Slice、map、slice 操作。
   - 结果：通过。
   - 标准答案：Day 8 goroutine/WaitGroup、Day 9 channel、Day 10 worker pool/任务队列、Day 11 HTTP handler、Day 12 JSON 编解码。

## 测试结果

- 4 题通过，1 题回答正确但实现有 bug。
- Bug：`BatchHandler` 中 `defer r.Body.Close()` 应移到解码之前，确保 JSON 解析失败时也能关闭请求体。

## 遇到的问题

- `ProcessTask` 输出字符串需要加空格：`"processed " + task.Data`，而不是 `"processed" + task.Data`。

## 关键收获

1. HTTP + JSON + 并发可以组合成完整的 API handler：接收请求、解析 JSON、并发处理、返回结果。
2. Worker pool 模式用 jobs channel 分发任务、results channel 收集结果、WaitGroup 等待 worker 退出。
3. 并发处理结果顺序不稳定，需要按业务字段（如 ID）排序后再返回或比较。
4. HTTP handler 需要处理方法检查、JSON 解析错误、响应头设置等边界情况。
5. Week 2 的核心能力已串联：goroutine、channel、任务队列、HTTP server、JSON 处理。

## 明日计划

- 进入 Day 15：Gin 框架和路由组。
