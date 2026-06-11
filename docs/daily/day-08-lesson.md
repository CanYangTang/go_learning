# Day 08 教案：goroutine 并发基础

## 学习目标

学完今天，需要能够做到：

1. 理解并发和并行的区别。
2. 使用 `go` 关键字启动 goroutine。
3. 理解主 goroutine 结束后，程序会直接退出。
4. 使用 `sync.WaitGroup` 等待多个 goroutine 完成。
5. 理解 `Add`、`Done`、`Wait` 的配合方式。
6. 通过测试验证并发执行的结果。
7. 对比串行执行和并发执行的耗时差异。

---

## Week 2 从并发开始

Week 1 学的是 Go 基础语法和工程结构。

Week 2 开始进入 Go 很重要的能力：并发。

Go 的并发模型很轻量，最核心的入口就是：

```go
go someFunction()
```

这会启动一个新的 goroutine，让函数和当前代码并发执行。

---

## 并发和并行的区别

这两个词容易混淆。

### 并发 concurrency

并发强调“同时处理多个任务”的结构。

即使只有一个 CPU 核，也可以通过调度让多个任务交替执行。

### 并行 parallelism

并行强调“多个任务真的在同一时刻执行”。

这通常需要多个 CPU 核。

简单理解：

```text
并发：我能同时管理多件事。
并行：我真的同时做多件事。
```

Go 的 goroutine 帮你表达并发，Go runtime 会根据情况调度它们是否并行执行。

---

## goroutine 是什么

goroutine 是 Go 中轻量级的并发执行单元。

普通函数调用：

```go
sayHello()
```

当前代码会等待 `sayHello` 执行完，再继续往下走。

goroutine 调用：

```go
go sayHello()
```

Go 会启动一个新的 goroutine 执行 `sayHello`，当前代码会继续往下走。

---

## 最小示例

```go
package main

import "fmt"

func main() {
    go fmt.Println("hello from goroutine")
    fmt.Println("hello from main")
}
```

这段代码不一定能看到 goroutine 的输出。

原因是：`main` 函数所在的 goroutine 结束后，整个程序就结束了，其他 goroutine 可能还没来得及执行。

---

## 主 goroutine 结束意味着程序退出

Go 程序启动时，会先运行 `main` goroutine。

如果 `main` goroutine 结束，程序就结束。

这意味着：

```go
go doWork()
```

不会自动让程序等待 `doWork` 完成。

所以我们需要一种等待机制。

今天使用的是：

```go
sync.WaitGroup
```

---

## sync.WaitGroup 是什么

`sync.WaitGroup` 用来等待一组 goroutine 完成。

基本用法：

```go
var wg sync.WaitGroup

wg.Add(1)
go func() {
    defer wg.Done()
    doWork()
}()

wg.Wait()
```

含义：

- `Add(1)`：告诉 WaitGroup，有 1 个任务要等。
- `Done()`：告诉 WaitGroup，一个任务完成了。
- `Wait()`：阻塞等待，直到计数变回 0。

---

## Add、Done、Wait 的关系

可以把 WaitGroup 想成一个计数器。

```go
wg.Add(3) // 计数器 +3
wg.Done() // 计数器 -1
wg.Done() // 计数器 -1
wg.Done() // 计数器 -1
wg.Wait() // 等到计数器变成 0
```

如果 `Done` 次数少于 `Add`，`Wait` 会一直等。

如果 `Done` 次数多于 `Add`，程序会 panic。

---

## 为什么常用 defer wg.Done()

推荐写法：

```go
go func() {
    defer wg.Done()
    doWork()
}()
```

这样可以保证当前 goroutine 返回前一定会调用 `Done`。

如果中间有多个 return，`defer wg.Done()` 也能执行。

这和 Day 5 学过的 `defer` 形成衔接。

---

## 循环中启动 goroutine

常见场景：对一组任务并发执行。

```go
for _, task := range tasks {
    wg.Add(1)
    go func(task string) {
        defer wg.Done()
        process(task)
    }(task)
}

wg.Wait()
```

注意：把 `task` 作为参数传给匿名函数。

这样每个 goroutine 使用的是当前循环对应的 task。

---

## 循环变量捕获问题

在循环里写 goroutine 时，容易写成：

```go
for _, task := range tasks {
    wg.Add(1)
    go func() {
        defer wg.Done()
        process(task)
    }()
}
```

在较老 Go 版本中，这可能导致 goroutine 捕获同一个循环变量，出现结果混乱。

更稳妥的写法是：

```go
for _, task := range tasks {
    task := task
    wg.Add(1)
    go func() {
        defer wg.Done()
        process(task)
    }()
}
```

或：

```go
for _, task := range tasks {
    wg.Add(1)
    go func(task string) {
        defer wg.Done()
        process(task)
    }(task)
}
```

今天建议使用第二种：参数传入。

---

## 并发结果的顺序不稳定

多个 goroutine 同时执行时，谁先完成不一定。

例如任务：

```text
A, B, C
```

并发执行结果可能是：

```text
B, A, C
```

也可能是：

```text
C, A, B
```

所以并发测试中不要随便依赖输出顺序。

如果确实需要稳定顺序，需要额外设计结果收集和排序逻辑。

---

## 数据竞争是什么

多个 goroutine 同时读写同一个变量，可能发生数据竞争。

例如：

```go
count := 0

for i := 0; i < 1000; i++ {
    go func() {
        count++
    }()
}
```

`count++` 不是原子操作，多个 goroutine 同时修改会出问题。

今天先不深入 mutex 和 channel，但要记住：

> goroutine 并发执行时，共享变量要特别小心。

今天的练习会尽量通过每个 goroutine 返回独立结果，避免共享写。

---

## 串行 vs 并发

串行执行：

```go
for _, task := range tasks {
    run(task)
}
```

一个任务完成后再执行下一个。

并发执行：

```go
for _, task := range tasks {
    wg.Add(1)
    go func(task Task) {
        defer wg.Done()
        run(task)
    }(task)
}
wg.Wait()
```

多个任务可以重叠执行。

如果每个任务都需要等待 100ms，3 个任务：

- 串行大约 300ms。
- 并发大约 100ms 多一点。

---

## time.Sleep 在练习中的作用

`time.Sleep` 可以模拟耗时任务。

```go
time.Sleep(100 * time.Millisecond)
```

真实项目中耗时可能来自：

- 网络请求
- 数据库查询
- 文件 IO
- 复杂计算

学习阶段用 `Sleep` 可以直观看到串行和并发的区别。

---

## 今日代码练习设计

今天建议创建：

```text
week02-core/day08-goroutine/
├── goroutine.go
└── goroutine_test.go
```

实现：

```go
type Task struct {
    Name     string
    Duration time.Duration
}

func RunSerial(tasks []Task) []string
func RunConcurrent(tasks []Task) []string
func MeasureSerial(tasks []Task) time.Duration
func MeasureConcurrent(tasks []Task) time.Duration
```

---

## Task 类型

```go
type Task struct {
    Name     string
    Duration time.Duration
}
```

表示一个任务：

- `Name`：任务名称。
- `Duration`：任务耗时。

执行任务时可以：

```go
time.Sleep(task.Duration)
return task.Name
```

---

## RunSerial

目标：串行执行任务。

```go
func RunSerial(tasks []Task) []string
```

规则：

1. 创建结果 slice。
2. 按顺序遍历 tasks。
3. 每个任务 sleep 对应时间。
4. 把任务名 append 到结果。
5. 返回结果。

串行结果顺序应该和输入顺序一致。

---

## RunConcurrent

目标：并发执行任务。

```go
func RunConcurrent(tasks []Task) []string
```

规则：

1. 使用 `sync.WaitGroup`。
2. 为每个任务启动 goroutine。
3. 每个 goroutine sleep 后记录任务名。
4. 等待所有 goroutine 完成。
5. 返回所有任务名。

为了避免共享 append 的数据竞争，建议创建固定长度结果 slice：

```go
results := make([]string, len(tasks))
```

每个 goroutine 只写自己的下标：

```go
results[index] = task.Name
```

这样能保留输入顺序，也减少共享 append 的风险。

---

## MeasureSerial 和 MeasureConcurrent

目标：测量耗时。

```go
func MeasureSerial(tasks []Task) time.Duration {
    start := time.Now()
    RunSerial(tasks)
    return time.Since(start)
}
```

并发版本类似：

```go
func MeasureConcurrent(tasks []Task) time.Duration {
    start := time.Now()
    RunConcurrent(tasks)
    return time.Since(start)
}
```

测试时不用要求精确时间，只要验证并发明显快于串行即可。

---

## 今日验收标准

完成后应该满足：

1. 创建 `week02-core/day08-goroutine/`。
2. 实现 `Task`、`RunSerial`、`RunConcurrent`、`MeasureSerial`、`MeasureConcurrent`。
3. 至少写 5 组测试。
4. `RunSerial` 保持输入顺序。
5. `RunConcurrent` 能等待所有 goroutine 完成。
6. 并发耗时明显小于串行耗时。
7. `go test ./...` 通过。
8. 能解释 goroutine、`sync.WaitGroup`、`Add`、`Done`、`Wait` 和循环变量捕获问题。

---

## 今天最容易踩的坑

### 坑 1：启动 goroutine 后主函数直接结束

```go
go doWork()
return
```

这样不保证 goroutine 执行完成。

应该用 `WaitGroup` 等待。

---

### 坑 2：忘记 Done

```go
wg.Add(1)
go func() {
    doWork()
}()
wg.Wait()
```

没有 `Done`，`Wait` 会一直等。

---

### 坑 3：Add 写在 goroutine 里面

不推荐：

```go
go func() {
    wg.Add(1)
    defer wg.Done()
    doWork()
}()
```

`Add` 应该在启动 goroutine 前调用，避免 `Wait` 提前看到计数为 0。

---

### 坑 4：并发 append 同一个 slice

多个 goroutine 同时 append 同一个 slice 可能数据竞争。

今天建议用固定长度 slice，每个 goroutine 写自己的下标。

---

### 坑 5：测试依赖并发完成顺序

并发完成顺序不稳定。

如果测试只关心“都完成了”，可以排序后比较，或使用固定下标保存结果。

今天建议固定下标保存，保持结果和输入顺序一致。
