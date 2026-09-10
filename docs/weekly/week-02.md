# Week 02 复盘：HTTP, JSON, Concurrency

## 时间范围

2026-06-11 ~ 2026-07-08

七天记录跨了近一个月：Day 8 / 9 / 10 都记在 2026-06-11，Day 11（06-12）到 Day 12（07-02）之间空了 20 天，Day 13 / 14 同为 07-08。

> 补记说明：本文件是 Day 21（2026-09-08）回补的。`GO_LEARN_PLAN.md` 里 Day 7 产出 `week-01.md`、Day 21 产出 `week-03.md`、Day 28 产出 `week-04.md`，没有任何一天负责 `week-02.md` —— 这是计划的漏洞，不是漏做的任务（`docs/issues-backlog.md` E1）。内容全部从 `docs/daily/day-08.md` ~ `day-14.md` 回收。

## 本周目标

- 掌握 goroutine 和 `sync.WaitGroup`，理解串行与并发的区别。
- 掌握 channel：无缓冲 / 有缓冲、`close` + `for range`、`select` 超时。
- 用 worker pool 实现一个任务队列，理解 producer-consumer。
- 用标准库 `net/http` 起一个 HTTP 服务，掌握 JSON 编解码。
- 掌握文件读写。
- 把并发和 HTTP 结合起来做一次综合练习。

## 已完成 Issue

- Day 8: goroutine 并发基础（issue #8）
- Day 9: channel 通信基础（issue #9）
- Day 10: 并发任务队列（issue #10）
- Day 11: HTTP Server 和 Health Endpoint（issue #11）
- Day 12: JSON 请求和响应处理（issue #12）
- Day 13: 文件操作（issue #13）
- Day 14: HTTP + 并发综合练习（issue #14）

（日志里当天只记了 issue 链接、没记标题，上面的名字取自各日教案的标题。）

## 本周代码产出

- Day 8：`week02-core/day08-goroutine/` —— `Task` + `RunSerial` / `RunConcurrent` / `MeasureSerial` / `MeasureConcurrent`。练 goroutine 启动、`WaitGroup` 等待、串并行耗时对比、用固定长度 slice 按下标写入避免竞争。
- Day 9：`week02-core/day09-channel/` —— `SendMessage` / `BufferedMessages` / `GenerateNumbers` / `ReceiveWithTimeout`。练无缓冲与有缓冲 channel、`close` + `for range`、`select` + `time.After`。
- Day 10：`week02-core/day10-task-queue/` —— `Job` / `Result` / `ProcessJobs` / `MeasureProcessJobs`。练 producer-consumer、jobs 与 results 两条 channel、多 worker、graceful shutdown。
- Day 11：`cmd/server/` —— 用 `http.NewServeMux()` 建独立路由器，注册 `GET /health` 返回 `{status, service, version}`，加 `main_test.go`。
- Day 12：`week02-core/day12-json/` —— `EchoRequest` / `EchoResponse` / `ErrorResponse` + `EchoHandler`。练 `json.NewDecoder(r.Body).Decode`、`Encoder` 写响应、struct tag、方法检查。
- Day 13：`week02-core/day13-files/` —— `ReadAll`（`os.ReadFile`）、`ReadLines`（`bufio.Scanner`）、`WriteAll`（`os.WriteFile`）、`WriteLines`（`bufio.Writer` + `Flush`）。
- Day 14：`week02-core/day14-summary/` —— `ProcessTask` / `ProcessBatch` / `BatchHandler`，把 Day 8 ~ 12 串起来：worker pool 并发处理 + `sort.Slice` 稳定排序 + HTTP/JSON 出入口。

## 掌握的概念

### goroutine 与 WaitGroup

- `go fn()` 并发执行，当前 goroutine 继续往下走；`fn()` 同步阻塞。
- 主 goroutine 结束进程就退出，不会等其他 goroutine。
- `sync.WaitGroup` 可直接 `var wg sync.WaitGroup`：`Add` 加计数、`Done` 减一、`Wait` 阻塞到归零。**`Add` 必须在启动 goroutine 之前调用**，否则 `Wait` 可能看到计数为 0 直接返回。
- 多个 goroutine 同时 `append` 同一个 slice 会并发修改 slice header 或底层数组，产生数据竞争。改成预先创建固定长度的 slice、每个 goroutine 只写自己的下标即可。
- `count++` 不是原子操作，可拆成读取、加一、写回。非原子操作的关键风险不是「看到半成品值」，而是多个 goroutine 在这几步之间交错执行。
- mutex 强调「同一时刻只有一个 goroutine 改数据」，channel 强调「通过消息传递协作」。

### channel

- 无缓冲 channel 的 send 阻塞到接收方就绪；有缓冲的在缓冲未满时可先写入。
- `value, ok := <-ch`：正常收到 `ok == true`；channel 已关闭**且缓冲已读空**时得到零值且 `ok == false`；已关闭但缓冲还有数据时，先读出缓冲数据且 `ok == true`。
- `close(ch)` 通常由发送方调用，用来通知接收方不会再有新数据。向已关闭的 channel 发送会 panic（`send on closed channel`）。
- `for range ch` 直到 channel 关闭**且**已发送的数据全部被接收完才结束。接收方用 `for range` 时发送方通常必须 close；接收方明确知道次数时可以不 close。
- `select` 同时等多个 channel 操作，谁先就绪执行谁。`time.After(d)` 接受一个 `time.Duration` 参数并**返回一个 channel**，经过 `d` 后该 channel 收到一个值 —— 把业务 channel 和它放进两个 case 就是超时控制。
- 超时返回后可能已经没有接收方了，此时后台 goroutine 再向无缓冲 channel 发送会永久阻塞，造成 goroutine 泄漏。把 channel 改成容量 1 的 buffered channel，后台 goroutine 完成一次发送就能正常退出。

### worker pool

- 至少要有 1 个 worker 消费 jobs channel，否则 producer 发送时没有接收方会阻塞 —— 所以 `workerCount <= 0` 要退化为 1。
- results channel 用容量 `len(jobs)` 的 buffered channel，降低 worker 发送结果被阻塞的概率。
- **收集结果和关闭 channel 必须并发**：主 goroutine 在 `for result := range results` 阻塞收集时，用 `go func() { wg.Wait(); close(results) }()` 单独等待并关闭。先 `wg.Wait()` 再 `close(results)` 会死锁。
- 关闭 results 必须等所有 worker 退出之后，否则 worker 再发送就 panic。
- 并发完成顺序不稳定，测试必须按业务字段排序或用 map 比较，不能依赖返回顺序。
- worker 不是越多越快：goroutine 虽轻量但有调度和栈内存成本，还会加剧对 channel、锁、数据库连接、下游服务的竞争。任务少而 worker 多时，大部分 worker 只是启动、等待、退出。实际应按 CPU 核数、任务类型、下游承载能力来定。

### net/http

- `http.DefaultServeMux` 是 `net/http` 预建的全局路由器，`http.HandleFunc` 注册到它，`ListenAndServe(addr, nil)` 也用它。`http.NewServeMux()` 创建独立路由器，不污染全局状态，测试和复杂项目更可控。
- `mux.HandleFunc("GET /health", h)` 是 Go 1.22+ 的 method-aware pattern。POST `/health` 返回 405 是 **`ServeMux` 依据路由模式自动处理的**，不是 handler 内部判断。
- `w http.ResponseWriter` 代表当前请求要写回客户端的响应，不是普通变量；`r *http.Request` 是客户端请求。
- `httptest.NewRecorder` 实现了 handler 需要的 `ResponseWriter` 行为，所以可以直接 `handler(recorder, req)` 或 `mux.ServeHTTP(recorder, req)`，**不需要真的监听 TCP 端口**。

### JSON

- `Marshal`/`Unmarshal` 在 `[]byte` 和 Go 值之间转换（Unmarshal 需传指针）；`Encoder.Encode` 直接把 JSON 写进 `io.Writer`，适合 HTTP response body 这类流式场景。
- struct tag `json:"message"` 指定字段名映射，不写时默认用 Go 字段名（大写形式）。
- `r.Body` 是 `io.ReadCloser`，实现了 `io.Reader`，所以能直接交给 `json.NewDecoder` 流式解析，比先全读进内存再 `Unmarshal` 更高效。
- `defer` 在函数返回时才执行，放函数开头只是「提前登记」，不影响中间读取 body。放开头的**关键原因是覆盖所有 early return 路径**（方法检查失败、解码失败），防止资源泄漏。
- `_ =` 是空白标识符，表示显式丢弃返回的 error。

### 文件、io 与类型

- struct 定义字段存数据，interface 定义方法签名不存数据；Go 的 interface 无需显式声明实现，方法签名匹配即自动满足 —— `*os.File` 因为实现了 `Read` 就自动满足 `io.Reader`。
- `os.ReadFile`/`os.WriteFile` 一步读写并自动关闭，适合小文件；`os.Open`/`os.Create` 返回 `*os.File`，需手动关闭，适合流式处理大文件。
- `bufio.Scanner`：`Scan()` 推进一行，EOF 或出错返回 `false`（出错用 `scanner.Err()` 区分）；`Text()` 返回当前行，不含换行符。
- `bufio.Writer` 必须 `Flush()`，忘记会导致内容丢失或不完整。
- 文件权限 `0644` = owner 读写、其他用户只读。
- 测试用 `os.CreateTemp("", "test*.txt")`（第一个参数 `""` 表示系统临时目录）配合 `defer os.Remove(file.Name())` 清理。
- `[]T` 是切片长度可变，`[N]T` 是数组长度固定不能 `append`。`append` 是函数不是方法，必须写 `lines = append(lines, line)`。
- `time.Duration` 底层单位是纳秒，`time.Millisecond` 是一个 `time.Duration` 常量 —— 已经是 duration 的值不能再乘单位常量。

## 仍然薄弱的点

- Day 10 是本周最弱的一天（2 题通过、2 题基本通过、1 题部分通过）：过早关闭 results channel 的**最关键**风险（worker 再发送会 panic）没答出来，`for range` 的准确结束条件也没答全。
- `time.Duration` 与单位常量的关系答不全（Day 8）。
- 对「原子操作」的理解需要更精确 —— 关键在多步骤被交错执行，不在中间值（Day 8）。
- `defer r.Body.Close()` 放在函数开头的关键原因（覆盖所有 early return）Day 12 没答出来，Day 14 概念答对但代码写错了位置。同一个点错了两次。
- Day 12 和 Day 14 的「答疑记录」都是「待记录」—— 这两天教案阶段没有留下任何提问。
- 各日的可选挑战题在产出和 review 里都没有完成记录，可视为未落地：Day 11「增加简单 request logging」、Day 12「`message` 非空且长度不超过 100 的校验」、Day 13「读取 JSON 文件并解析成 struct」、Day 14「增加 context timeout 或 cancellation」。（这是从「产出未提及」推断的，日志没有显式写「未完成」。）
- `context` 整个概念本周没碰到，超时控制全靠 `time.After`。
- 本周所有「`make test` 通过」的结论都受一个缺口影响：`make test` 是裸 `go test ./...`，没有 `-count=1`，Day 20 已经被测试缓存骗过一次（`docs/issues-backlog.md` C3）。

## 典型错误与修正

- Day 8：把已经是 `time.Duration` 的 `task.Duration` 又乘了一次 `time.Millisecond`，30ms 被放大成约 8.33 小时 → duration 值直接传给 `time.Sleep`。
- Day 9：`GenerateNumbers` 生成了 `0..n-1`，测试期望 `1..n` → 改为 `1..n`。
- Day 9：`ReceiveWithTimeout` 初版用无缓冲 channel，超时返回后后台 goroutine 永久阻塞（goroutine 泄漏）→ 改为 `make(chan string, 1)`。
- Day 9（小测）：把 `time.After` 说成「可以接收一个 `time.Duration` 值」→ 它是**接受** `time.Duration` 参数并**返回一个 channel**。
- Day 10（小测）：只说过早关闭 results 会「结果不完整」→ 关键后果是 worker 再发送触发 `send on closed channel` panic。
- Day 10（小测）：说测试没有处理顺序不稳定 → 测试其实已经按 `JobID` 排序了。
- Day 11（小测）：说 POST 405 是在 `healthHandler` 里判断的 → 是 `ServeMux` 依据 method-aware pattern 自动处理的。
- Day 11（小测）：「为什么测试 handler 不需要真的启动端口还不清楚」→ `httptest.NewRecorder` 实现了 `ResponseWriter`，可直接调用 handler。
- Day 11：混淆了 `DefaultServeMux`（标准库的全局路由器）和 `http.NewServeMux()`（独立路由器）。
- Day 12（小测）：只说出 `defer` 是「提前登记」，漏了「覆盖所有 early return 路径」这个关键原因。
- Day 13：`lines := []` 不合法，初始化 slice 必须带元素类型；`append` 是函数不是方法；`WriteLines` 每行要补 `"\n"`。Day 13 小测 5 题全对，是本周唯一满分。
- Day 14：`ProcessTask` 输出字符串漏了空格，应为 `"processed " + task.Data`。
- Day 14（小测第 4 题）：概念答对但实现有 bug —— `BatchHandler` 的 `defer r.Body.Close()` 写在解码之后，解码失败 return 时 defer 还没登记，body 不会关闭。**这个 bug 后来已修**（现在在方法检查之后、解码之前），但日志里没有记录修复过程。
- Day 12 的日志里写着「`defer r.Body.Close()` 放在解码成功后可以确保正常流程关闭请求体」，与同日小测的标准答案（应放开头）自相矛盾；这条在 Day 14 被推翻。

小测通过率：Day 8「4 通过 + 1 需补细节」、Day 9「4 通过 + 1 需修表述」、Day 10「2 通过 + 2 基本通过 + 1 部分通过」、Day 11「3 通过 + 2 基本通过」、Day 12「4 通过 + 1 基本通过」、Day 13「5 全对」、Day 14「4 通过 + 1 答对但实现有 bug」。

## 代码 Review 结论

- Day 8：`Task` 用 `time.Duration` 表达耗时，与 `time.Sleep` 参数类型一致；`RunConcurrent` 的 `Add` 在启动 goroutine 前、`Done` 用 `defer`；结果用固定长度 slice 按下标写入，避免并发 append；循环变量 `i` 和 `task` 作为参数传给匿名函数，避免捕获问题。
- Day 9：四个函数分别覆盖无缓冲同步交接、有缓冲批量发送、`close` + `for range`、`select` 超时；超时那个用容量 1 的 channel 避免泄漏。
- Day 10：`workerCount <= 0` 退化为 1；producer 发完就 `close(jobs)` 让 worker 从 `for range` 自然退出；`wg.Wait()` 后才关 results；测试按 `JobID` 排序后比较。
- Day 11：用 `http.NewServeMux()` 不依赖全局路由器；method-aware pattern 让非 GET 自动 405；`ListenAndServe(addr, mux)` 显式传路由器；测试用 `httptest` 直接测 handler。
- Day 12：方法检查在最前面，结构清晰；用 `Decoder` 流式解码；成功时设 `Content-Type` 返回 200。**两处待改**：错误信息硬编码成 `"..."`；`defer r.Body.Close()` 的位置当时被判为「正确」，Day 14 的标准答案推翻了这个结论。
- Day 13：`ReadAll`/`WriteAll` 用一步式 API 适合小文件，`ReadLines`/`WriteLines` 用 `bufio` 流式处理；`defer file.Close()` 在打开成功后；测试用临时文件并清理。
- Day 14：复用 Day 10 的 worker pool 模式，`sort.Slice` 按 `ID` 稳定排序，HTTP 层完成方法检查 / 解析 / 响应。

## 验证命令

```bash
# 逐日包级测试
go test ./week02-core/day08-goroutine
go test ./week02-core/day09-channel
go test ./week02-core/day10-task-queue
go test ./cmd/server
go test ./week02-core/day12-json
go test ./week02-core/day13-files
go test ./week02-core/day14-summary

# 每日固定组合
make fmt
make test
make vet

# Day 8 ~ 11 另外记录
go list ./...

# 仅 Day 11
make run
curl http://localhost:8080/health
```

Day 8 / 9 / 10 的验收标准里写了 `go test ./...`，但「运行过的命令」里从未出现过这条原样命令，实际跑的是 `make test`。

## 下周优先级

1. 从 `net/http` 切到 Gin，建立路由组和统一 JSON 响应。
2. 打通 MySQL：连接、迁移、CRUD，再升级到 GORM。
3. 补上中间件层，并把代码重构成 handler / service / repository 三层。


