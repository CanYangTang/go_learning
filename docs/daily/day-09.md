# Day 09 学习记录

## 日期

2026-06-11

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/9

## 今日教案

- 教案文档：`docs/daily/day-09-lesson.md`

## 核心任务

- 学习 channel 通信基础。
- 使用无缓冲和有缓冲 channel 在 goroutine 之间传递数据。
- 使用 `close`、`for range`、`select` 和 `time.After` 完成基础同步和超时控制。

## 验收标准

- 创建 `week02-core/day09-channel/`。
- 实现 `SendMessage`、`BufferedMessages`、`GenerateNumbers`、`ReceiveWithTimeout`。
- 至少写 6 组测试。
- `go test ./week02-core/day09-channel` 通过。
- `go test ./...` 通过。
- 能解释无缓冲 channel、有缓冲 channel、`close`、`for range`、`select`、`time.After`。

## 可选挑战题

- 使用 `select` 增加超时控制。

## 学习前问题

- `value, ok := <-ch` 如果 channel 已关闭，但还有缓冲数据，会返回什么？

## 答疑记录

- 如果 channel 已关闭但缓冲区里还有数据，接收时会继续读出缓冲数据，并且 `ok == true`。
- 只有当 channel 已关闭且缓冲区数据也被读完后，再接收才会得到零值，并且 `ok == false`。
- `for range ch` 会持续读取 channel 中的数据，直到 channel 关闭且缓冲区数据读完后自然结束。

## 今日产出

- 创建 Day 9 练习目录：`week02-core/day09-channel/`。
- 实现 `SendMessage`、`BufferedMessages`、`GenerateNumbers`、`ReceiveWithTimeout`。
- 编写测试覆盖无缓冲 channel、有缓冲 channel、`close` + `for range`、成功接收和超时分支。

## 运行过的命令

```bash
go test ./week02-core/day09-channel
go list ./...
make fmt
make test
make vet
```

## 代码 Review 结论

- `SendMessage` 使用无缓冲 channel，在 goroutine 中发送消息，主 goroutine 接收消息，体现了 channel 的同步交接特性。
- `BufferedMessages` 使用容量为 `len(messages)` 的 buffered channel，能先发送多条消息再按数量接收。
- `GenerateNumbers` 由发送方在发送完 `1..n` 后关闭 channel，接收方使用 `for range` 读取直到 channel 关闭。
- `ReceiveWithTimeout` 使用 `select` 同时等待消息和 `time.After(timeout)`，能处理成功接收和超时场景。
- `ReceiveWithTimeout` 的消息 channel 使用容量为 1 的 buffered channel，避免超时返回后后台 goroutine 发送消息时永久阻塞。
- `go test ./week02-core/day09-channel`、`make fmt`、`make test`、`make vet` 均已通过。

## 今日小测试

1. 无缓冲 channel 和有缓冲 channel 的发送行为有什么区别？
   - 回答：无缓冲 channel 的发送方会阻塞，直到有接收方接收；有缓冲 channel 在缓冲未满时，可以先放入。
   - 结果：通过。
   - 标准答案：无缓冲 channel 发送时必须等接收方准备好；有缓冲 channel 在缓冲区未满时可以先写入，缓冲区满了才会阻塞。
2. `value, ok := <-ch` 中的 `ok` 什么时候是 `true`，什么时候是 `false`？如果 channel 已关闭但缓冲区还有数据，`ok` 是什么？
   - 回答：有数据时为 true，无数据且 channel 已关闭为 false；已关闭但缓冲区还有数据时为 true。
   - 结果：通过。
   - 标准答案：正常接收到数据时 `ok == true`；channel 已关闭且缓冲区数据已读完时，接收到零值且 `ok == false`。如果 channel 已关闭但缓冲区还有数据，会先读出缓冲数据，`ok == true`。
3. 为什么 `for value := range ch` 通常需要发送方在合适的时候 `close(ch)`？
   - 回答：因为接收方不知道发送方什么时候会停止发数据，所以会一直循环接收。
   - 结果：通过。
   - 标准答案：`for range ch` 会持续接收，直到 channel 被关闭且缓冲区数据读完。如果发送方不关闭 channel，接收方可能一直阻塞等待下一条数据。
4. `select` 和 `time.After` 如何配合实现超时控制？
   - 回答：利用 `select` 可以同时等待多个 channel 操作的特性，哪个先准备好，就执行哪个；`<-time.After(d)` 可以接收一个 `time.Duration` 值，过了指定时间，channel 就会收到一个值。
   - 结果：基本通过，需要修正表述。
   - 标准答案：`select` 可以同时等待多个 channel 操作；`time.After(d)` 接收一个 `time.Duration` 参数，并返回一个 channel，经过时间 `d` 后该 channel 会收到一个值。把业务 channel 和 `time.After(timeout)` 放在两个 case 中，谁先准备好就执行谁。
5. 为什么 `ReceiveWithTimeout` 中用于接收消息的 channel 最终改成了 `make(chan string, 1)`？它解决了什么问题？
   - 回答：`messages := make(chan string)` 是无缓冲 channel。超时时，`ReceiveWithTimeout` 已经返回了，但后台 goroutine 之后还会执行 `messages <- message`。因为没有接收方了，它会永远阻塞，形成 goroutine 泄漏。改完后，即使外层已经因为 timeout 返回，后台 goroutine 发送一次也能放进缓冲区并结束。
   - 结果：通过。
   - 标准答案：无缓冲 channel 在超时返回后可能没有接收方，后台 goroutine 再发送会永久阻塞。容量为 1 的 buffered channel 可以让后台 goroutine 在超时后仍然完成一次发送并退出，避免 goroutine 泄漏。

## 测试结果

- 4 题通过，1 题基本通过但需要修正表述。
- 需要修正：`time.After(d)` 是接收一个 `time.Duration` 参数并返回 channel，不是接收一个 `time.Duration` 值。

## 遇到的问题

- 初次实现 `GenerateNumbers` 时生成的是 `0..n-1`，测试期望是 `1..n`。
- `ReceiveWithTimeout` 初次使用无缓冲 channel，超时返回后后台 goroutine 可能因为没有接收方而永久阻塞；改为容量为 1 的 buffered channel 后，发送方可以发送一次并正常退出。
- 对 `close` 的作用需要区分：如果接收方使用 `for range ch`，通常需要发送方关闭 channel 来表示不会再发送；如果接收方明确知道接收次数，可以不关闭。

## 关键收获

1. channel 是 goroutine 之间传递数据和同步信号的管道。
2. 无缓冲 channel 发送和接收必须同步配对，有缓冲 channel 可以在缓冲未满时暂存数据。
3. `close(ch)` 通常由发送方调用，用来通知接收方不会再有新数据。
4. `for range ch` 会持续接收，直到 channel 关闭且缓冲数据读完。
5. `select` 可以同时等待多个 channel 操作，配合 `time.After` 可以实现超时控制。

## 明日计划

- 进入 Day 10：并发任务队列。
- 练习 worker、任务分发、结果收集和并发控制。
