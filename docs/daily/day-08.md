# Day 08 学习记录

## 日期

2026-06-11

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/8

## 今日教案

- 教案文档：`docs/daily/day-08-lesson.md`

## 核心任务

- 学习 goroutine 并发基础。
- 使用 `sync.WaitGroup` 等待多个 goroutine 完成。
- 对比串行和并发执行耗时。

## 验收标准

- 创建 `week02-core/day08-goroutine/`。
- 实现 `Task`、`RunSerial`、`RunConcurrent`、`MeasureSerial`、`MeasureConcurrent`。
- 至少写 5 组测试。
- `RunSerial` 保持输入顺序。
- `RunConcurrent` 能等待所有 goroutine 完成。
- 并发耗时明显小于串行耗时。
- `go test ./...` 通过。
- 能解释 goroutine、`sync.WaitGroup`、`Add`、`Done`、`Wait` 和循环变量捕获问题。

## 可选挑战题

- 对比串行执行和并发执行耗时。

## 学习前问题

- `sync.WaitGroup` 是一种新的变量类型吗？
- 什么是原子操作？
- 非原子操作是不是意味着中间状态会被其他 goroutine 看到？
- `mutex` 和 `channel` 是什么？

## 答疑记录

- `sync.WaitGroup` 是标准库 `sync` 包里定义的结构体类型，可以声明变量使用，例如 `var wg sync.WaitGroup`，用于等待一组 goroutine 完成。
- 原子操作是不可分割的操作，要么完整完成，要么完全没发生；`count++` 通常不是原子操作，因为它可以拆成读取、加一、写回多个步骤。
- 非原子操作不一定表示其他 goroutine 会看到“半成品值”，更准确地说，是多个 goroutine 可能在读、改、写这些步骤之间交错执行，导致结果不符合预期。
- `mutex` 是互斥锁，常用于保护共享变量，同一时间只允许一个 goroutine 修改受保护的数据。
- `channel` 是 goroutine 之间通信的管道，一个 goroutine 可以发送数据，另一个 goroutine 可以接收数据；它更强调通过消息传递来协作。

## 今日产出

- 创建 Day 8 练习目录：`week02-core/day08-goroutine/`。
- 实现 `Task`、`RunSerial`、`RunConcurrent`、`MeasureSerial`、`MeasureConcurrent`。
- 编写测试覆盖串行执行顺序、空任务、并发等待完成和串行/并发耗时对比。

## 运行过的命令

```bash
go test ./week02-core/day08-goroutine
go list ./...
make fmt
make test
make vet
```

## 代码 Review 结论

- `Task` 使用 `time.Duration` 表达耗时，和 `time.Sleep` 的参数类型一致。
- `RunSerial` 按输入顺序逐个执行任务，结果顺序稳定。
- `RunConcurrent` 使用 `sync.WaitGroup` 等待所有 goroutine 完成，`Add` 在启动 goroutine 前调用，`Done` 使用 `defer` 保证执行。
- `RunConcurrent` 使用固定长度结果 slice，并让每个 goroutine 写入自己的下标，避免并发 append 同一个 slice。
- 循环中把 `i` 和 `task` 作为参数传给匿名函数，避免循环变量捕获问题。
- `MeasureSerial` 和 `MeasureConcurrent` 使用 `time.Now` 与 `time.Since` 测量耗时，能直观看出并发执行优势。
- `go test ./week02-core/day08-goroutine`、`make fmt`、`make test`、`make vet` 均已通过。

## 今日小测试

1. goroutine 是什么？`go fn()` 和普通的 `fn()` 调用有什么区别？
   - 回答：是 Go 里轻量级的并发执行单元；`go fn()` 是并发启动、异步执行，`fn()` 是同步执行。
   - 结果：通过。
   - 标准答案：goroutine 是 Go 的轻量级并发执行单元。普通 `fn()` 调用会阻塞当前流程，等函数执行完再继续；`go fn()` 会启动新的 goroutine 并发执行，当前 goroutine 会继续往下走。
2. 为什么主 goroutine 结束后，其他 goroutine 可能还没执行完程序就退出？今天用什么机制解决这个问题？
   - 回答：主 goroutine 结束后，当前程序会结束，其他 goroutine 也会被直接结束；用 `sync.WaitGroup`。
   - 结果：通过。
   - 标准答案：Go 程序由主 goroutine 驱动，主 goroutine 结束后进程会退出，不会自动等待其他 goroutine。今天使用 `sync.WaitGroup` 等待一组 goroutine 完成。
3. `sync.WaitGroup` 的 `Add`、`Done`、`Wait` 分别做什么？为什么 `Add` 推荐在启动 goroutine 前调用？
   - 回答：有多少任务要等、一个任务完成了、阻塞等待直到计数变回 0；`Add` 要先于 `go`，否则 `Wait` 可能在任务登记前就认为“没有任务可等”，直接结束。
   - 结果：通过。
   - 标准答案：`Add` 增加等待计数，`Done` 表示一个任务完成并让计数减一，`Wait` 阻塞直到计数归零。`Add` 应在启动 goroutine 前调用，避免 `Wait` 提前看到计数为 0。
4. 为什么多个 goroutine 同时 `append` 同一个 slice 可能有问题？今天的 `RunConcurrent` 是怎么避免这个问题的？
   - 回答：竞态；将 slice 的长度和 goroutine 保存一致，通过索引准确赋值。
   - 结果：通过。
   - 标准答案：多个 goroutine 同时 `append` 同一个 slice 可能并发修改 slice header 或底层数组，产生数据竞争。今天先创建固定长度 `results`，每个 goroutine 只写自己的下标，避免并发 append。
5. `time.Duration` 和 `time.Millisecond` 是什么关系？为什么 `task.Duration` 已经是 `30 * time.Millisecond` 时，不能再写 `task.Duration * time.Millisecond`？
   - 回答：类型和值的关系；会放大，造成预期外的影响。
   - 结果：基本通过，需要补充细节。
   - 标准答案：`time.Duration` 是表示时间长度的类型，底层单位是纳秒；`time.Millisecond` 是一个 `time.Duration` 常量，表示 1 毫秒。`30 * time.Millisecond` 已经得到一个 duration，如果再乘 `time.Millisecond`，会把底层纳秒数再次乘以 1,000,000，导致耗时被异常放大。

## 测试结果

- 4 题通过，1 题基本通过但需要补充细节。
- 需要补充：`time.Millisecond` 是 `time.Duration` 常量，`time.Duration` 底层以纳秒计数，重复乘单位会放大数值。

## 遇到的问题

- 初次实现时把已经是 `time.Duration` 的 `task.Duration` 又乘了一次 `time.Millisecond`，导致 30ms 被放大成约 8.33 小时。
- 对“原子操作”和“非原子操作”的理解需要更精确：非原子操作的关键问题是多个步骤之间可能被其他 goroutine 交错执行。

## 关键收获

1. goroutine 是 Go 的轻量级并发执行单元，使用 `go` 关键字启动。
2. `sync.WaitGroup` 可以等待一组 goroutine 完成，核心方法是 `Add`、`Done` 和 `Wait`。
3. 并发执行时要谨慎处理共享变量，避免多个 goroutine 同时修改同一个数据结构。
4. 固定长度 slice 加下标写入可以避免多个 goroutine 并发 append 同一个 slice。
5. `time.Duration` 本身表示时间长度，传给 `time.Sleep` 时不应该重复乘单位常量。

## 明日计划

- 进入 Day 9：channel 通信基础。
- 学习 channel 创建、发送、接收、关闭和基础同步。
