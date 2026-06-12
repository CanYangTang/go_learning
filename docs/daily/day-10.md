# Day 10 学习记录

## 日期

2026-06-11

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/10

## 今日教案

- 教案文档：`docs/daily/day-10-lesson.md`

## 核心任务

- 学习 producer-consumer 并发任务队列模型。
- 使用 jobs channel 分发任务，使用 results channel 收集结果。
- 使用多个 worker 并发处理任务，并通过 `sync.WaitGroup` 实现 graceful shutdown。

## 验收标准

- 创建 `week02-core/day10-task-queue/`。
- 实现 `Job`、`Result`、`ProcessJobs`、`MeasureProcessJobs`。
- 至少写 5 组测试。
- `workerCount <= 0` 时能退化为 1 个 worker。
- 能正确处理空任务。
- 多 worker 处理结果正确。
- 多 worker 耗时明显少于单 worker。
- `go test ./...` 通过。
- 能解释 producer-consumer、jobs channel、results channel、worker、graceful shutdown。

## 可选挑战题

- 增加可配置 worker 数量和 graceful shutdown。

## 学习前问题

- worker 太多为什么会造成资源浪费？

## 答疑记录

- worker 本质上是 goroutine，虽然轻量但不是零成本；数量太多会增加 Go runtime 的调度成本和 goroutine 栈内存成本。
- worker 太多还会增加对 jobs channel、results channel、锁、数据库连接、网络服务等共享资源或下游系统的竞争。
- 如果任务数量少而 worker 数量很多，大部分 worker 只是启动、等待、退出，并没有实际提升吞吐。
- worker 数量的作用是控制并发度，不是越多越快，实际工程中通常要根据 CPU 核数、任务类型和下游承载能力设置。

## 今日产出

- 创建 Day 10 练习目录：`week02-core/day10-task-queue/`。
- 实现 `Job`、`Result`、`ProcessJobs`、`MeasureProcessJobs`。
- 编写测试覆盖多 worker 处理、空任务、`workerCount <= 0` 兜底、worker 数量多于任务数和耗时对比。

## 运行过的命令

```bash
go test ./week02-core/day10-task-queue
go list ./...
make fmt
make test
make vet
```

## 代码 Review 结论

- `ProcessJobs` 在 `workerCount <= 0` 时退化为 1 个 worker，避免无人消费 jobs channel 导致发送阻塞。
- jobs channel 用于分发任务，多个 worker 通过 `for job := range jobsChan` 持续消费任务。
- results channel 使用容量为 `len(jobs)` 的 buffered channel，降低 worker 发送结果时被阻塞的概率。
- producer 发送完所有任务后关闭 jobs channel，让 worker 能从 `for range` 中自然退出。
- 使用 `sync.WaitGroup` 等待所有 worker 退出，再关闭 results channel，避免 worker 仍在发送时关闭 channel。
- 测试中对结果按 `JobID` 排序后比较，避免依赖并发完成顺序。
- `go test ./week02-core/day10-task-queue`、`make fmt`、`make test`、`make vet` 均已通过。

## 今日小测试

1. producer-consumer 模型里 producer、jobs channel、worker、results channel 分别负责什么？
   - 回答：producer 负责生产任务；jobs channel 负责把任务放进队列；worker 负责从 jobs channel 接收任务并处理任务；results channel 负责接收 worker 处理任务后的结果。
   - 结果：通过。
   - 标准答案：producer 负责发送任务；jobs channel 负责在 producer 和 worker 之间传递任务；worker 负责从 jobs channel 接收任务并处理；results channel 负责把 worker 处理后的结果传回收集方。
2. 为什么 `workerCount <= 0` 时要退化为 1 个 worker？如果 worker 数量是 0 会发生什么？
   - 回答：必须保障有一个 worker 才能消费生产的任务；否则主 goroutine 会阻塞。
   - 结果：通过。
   - 标准答案：至少要有 1 个 worker 消费 jobs channel，否则 producer 向 jobs channel 发送任务时没有接收方，会阻塞，程序无法继续完成任务处理。
3. 为什么 producer 发送完任务后要 `close(jobsChan)`？worker 的 `for job := range jobsChan` 什么时候结束？
   - 回答：因为负责接收的 worker 得知道什么时候没有更多任务可以退出；主 goroutine `close(jobsChan)` 后结束。
   - 结果：基本通过，需要补充结束条件。
   - 标准答案：关闭 jobs channel 是为了通知 worker 不会再有新任务。`for job := range jobsChan` 会在 jobs channel 已关闭，并且 channel 中已发送的任务都被接收完之后结束。
4. 为什么并发任务队列里的结果顺序不一定和输入顺序一致？今天的测试是怎么处理这个问题的？
   - 回答：因为哪个 worker 先处理完任务是不确定的；今天的测试没有处理这个问题。
   - 结果：前半部分通过，后半部分需要修正。
   - 标准答案：多个 worker 并发处理任务，不同任务的耗时和调度顺序不固定，所以结果返回顺序不稳定。今天的测试在比较前按 `JobID` 排序，避免依赖并发完成顺序。
5. 为什么要等 `wg.Wait()` 之后再关闭 `resultsChan`？如果过早关闭 results channel 会发生什么？
   - 回答：`wg.Wait()` 后才代表所有 worker 都处理完了，也就是所有 job 都处理完了；如果提前关闭 `resultsChan`，那么收集到的结果将不完整，不能包含所有 job 的结果。
   - 结果：基本通过，需要补充最关键风险。
   - 标准答案：`wg.Wait()` 返回后说明所有 worker 都已经退出，不会再向 `resultsChan` 发送结果，此时关闭 results channel 才安全。如果过早关闭，worker 之后再发送结果会触发 `send on closed channel` panic。

## 测试结果

- 2 题通过，2 题基本通过，1 题部分通过。
- 需要修正：测试已经通过按 `JobID` 排序处理结果顺序不稳定的问题。
- 需要补充：过早关闭 `resultsChan` 的关键风险是 worker 再发送时会 panic。

## 遇到的问题

- worker 数量不是越多越好；worker 太多会增加调度、内存和共享资源竞争成本。
- 并发结果完成顺序不稳定，测试不能直接依赖返回顺序，需要排序或用 map 判断。

## 关键收获

1. producer-consumer 模型可以用 jobs channel 分发任务，用 worker 并发处理，用 results channel 收集结果。
2. `close(jobsChan)` 是通知 worker 没有更多任务的关键步骤，能让 `for range jobsChan` 自然结束。
3. `sync.WaitGroup` 可以等待所有 worker 退出，是 graceful shutdown 的基础组成部分。
4. results channel 只有在确认不会再有 worker 发送结果后才能关闭。
5. worker 数量用于控制并发度，需要根据任务数量、任务类型和下游承载能力设置。

## 明日计划

- 进入 Day 11：HTTP server 和 health endpoint。
- 练习 `net/http`、路由注册、请求处理和基础测试。
