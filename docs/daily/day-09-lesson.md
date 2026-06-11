# Day 09 教案：channel 通信基础

## 学习目标

学完今天，需要能够做到：

1. 理解 channel 是 goroutine 之间通信的管道。
2. 使用 `make(chan T)` 创建无缓冲 channel。
3. 使用 `ch <- value` 发送数据。
4. 使用 `value := <-ch` 接收数据。
5. 理解无缓冲 channel 的同步特性。
6. 使用 buffered channel 暂存有限数量的数据。
7. 使用 `close(ch)` 关闭 channel。
8. 使用 `for range` 遍历 channel 中的数据。
9. 使用 `select` 和 `time.After` 实现超时控制。

---

## 从 goroutine 到 channel

Day 8 学了 goroutine 和 `sync.WaitGroup`。

`WaitGroup` 解决的问题是：

```text
等待一组 goroutine 完成
```

但它不负责传递数据。

如果 goroutine 之间需要传递结果、任务或信号，就需要 channel。

channel 可以理解为：

```text
goroutine 之间传递数据的管道
```

---

## 创建 channel

使用 `make` 创建 channel：

```go
ch := make(chan int)
```

这里：

- `chan int` 表示这是一个传递 `int` 的 channel。
- `ch` 是 channel 变量。

字符串 channel：

```go
messages := make(chan string)
```

任务 channel：

```go
tasks := make(chan Task)
```

---

## 发送和接收

发送数据：

```go
ch <- 10
```

接收数据：

```go
value := <-ch
```

完整例子：

```go
ch := make(chan string)

go func() {
    ch <- "hello"
}()

msg := <-ch
fmt.Println(msg)
```

一个 goroutine 发送，另一个 goroutine 接收。

---

## 无缓冲 channel

默认创建的是无缓冲 channel：

```go
ch := make(chan string)
```

无缓冲 channel 的特点：

- 发送方会阻塞，直到有接收方接收。
- 接收方会阻塞，直到有发送方发送。

所以无缓冲 channel 既能传数据，也能做同步。

---

## 无缓冲 channel 的同步特性

例子：

```go
ch := make(chan string)

go func() {
    ch <- "done"
}()

msg := <-ch
```

执行到 `msg := <-ch` 时，主 goroutine 会等待。

等子 goroutine 发送 `"done"` 后，主 goroutine 才能继续执行。

这就是 channel 的同步能力。

---

## buffered channel

带缓冲的 channel：

```go
ch := make(chan string, 2)
```

第二个参数 `2` 表示缓冲区容量。

可以先放入最多 2 个值，而不需要立即有接收方：

```go
ch <- "a"
ch <- "b"
```

如果再发送第三个值：

```go
ch <- "c"
```

当缓冲区满了，又没有接收方时，发送方会阻塞。

---

## 无缓冲和有缓冲的区别

| 类型 | 创建方式 | 发送行为 |
|---|---|---|
| 无缓冲 | `make(chan string)` | 必须等接收方准备好 |
| 有缓冲 | `make(chan string, 2)` | 缓冲未满时可以先放入 |

简单理解：

- 无缓冲 channel 像直接交接物品。
- 有缓冲 channel 像有一个容量有限的邮箱。

---

## close(channel)

发送方可以关闭 channel：

```go
close(ch)
```

关闭 channel 表示：不会再有新数据发送进来。

注意：

- 通常由发送方关闭 channel。
- 不要向已关闭的 channel 发送数据，否则会 panic。
- 接收方可以继续从已关闭 channel 中接收已缓冲的数据。

---

## 接收时判断 channel 是否关闭

可以使用双返回值：

```go
value, ok := <-ch
```

含义：

- `ok == true`：接收到了正常发送的值。
- `ok == false`：channel 已关闭且没有剩余数据。

示例：

```go
value, ok := <-ch
if !ok {
    fmt.Println("channel closed")
}
_ = value
```

---

## for range 遍历 channel

如果发送方会关闭 channel，接收方可以用 `for range` 遍历：

```go
for value := range ch {
    fmt.Println(value)
}
```

这个循环会一直接收数据，直到 channel 被关闭。

所以使用 `for range ch` 时，要确保某处会调用 `close(ch)`，否则可能一直阻塞。

---

## select

`select` 可以同时等待多个 channel 操作。

```go
select {
case msg := <-messages:
    fmt.Println(msg)
case <-time.After(time.Second):
    fmt.Println("timeout")
}
```

哪个 case 先准备好，就执行哪个。

如果都没准备好，`select` 会阻塞。

---

## time.After 超时控制

`time.After(d)` 会返回一个 channel。

过了指定时间后，这个 channel 会收到一个值。

常用于超时控制：

```go
select {
case result := <-ch:
    return result, true
case <-time.After(100 * time.Millisecond):
    return "", false
}
```

含义：

- 如果 `ch` 先收到数据，返回结果。
- 如果 100ms 先到，返回超时。

---

## channel 和 WaitGroup 的区别

`WaitGroup` 适合：

```text
等所有 goroutine 完成
```

channel 适合：

```text
在 goroutine 之间传递数据、结果或信号
```

有时两者会一起用，但今天重点练 channel。

---

## 今日代码练习设计

今天创建：

```text
week02-core/day09-channel/
├── channel.go
└── channel_test.go
```

建议实现：

```go
func SendMessage(message string) string
func BufferedMessages(messages []string) []string
func GenerateNumbers(n int) []int
func ReceiveWithTimeout(message string, delay time.Duration, timeout time.Duration) (string, bool)
```

---

## SendMessage

目标：用无缓冲 channel 从 goroutine 发送一条消息。

```go
func SendMessage(message string) string
```

期望：

```go
SendMessage("hello") // "hello"
```

练习点：

- `make(chan string)`
- goroutine
- `ch <- message`
- `<-ch`

---

## BufferedMessages

目标：使用 buffered channel 暂存多条消息，并按顺序读出。

```go
func BufferedMessages(messages []string) []string
```

期望：

```go
BufferedMessages([]string{"a", "b"}) // []string{"a", "b"}
```

练习点：

- `make(chan string, len(messages))`
- buffered channel
- 发送多个值
- 接收多个值

---

## GenerateNumbers

目标：用 goroutine 向 channel 发送 `1..n`，关闭 channel 后用 `for range` 收集结果。

```go
func GenerateNumbers(n int) []int
```

期望：

```go
GenerateNumbers(3) // []int{1, 2, 3}
GenerateNumbers(0) // []int{}
```

练习点：

- `chan int`
- goroutine 中发送多个值
- `close(ch)`
- `for value := range ch`

---

## ReceiveWithTimeout

目标：使用 `select` 和 `time.After` 实现超时控制。

```go
func ReceiveWithTimeout(message string, delay time.Duration, timeout time.Duration) (string, bool)
```

规则：

- 启动 goroutine。
- goroutine 等待 `delay` 后发送 message。
- 主 goroutine 使用 `select` 等待消息或超时。
- 如果先收到消息，返回 `message, true`。
- 如果先超时，返回 `"", false`。

期望：

```go
ReceiveWithTimeout("ok", 10*time.Millisecond, 50*time.Millisecond) // "ok", true
ReceiveWithTimeout("slow", 50*time.Millisecond, 10*time.Millisecond) // "", false
```

练习点：

- `select`
- `time.After`
- 超时控制
- 多返回值

---

## 今日验收标准

完成后应该满足：

1. 创建 `week02-core/day09-channel/`。
2. 实现 `SendMessage`、`BufferedMessages`、`GenerateNumbers`、`ReceiveWithTimeout`。
3. 至少写 6 组测试。
4. `go test ./week02-core/day09-channel` 通过。
5. `go test ./...` 通过。
6. 能解释无缓冲 channel、有缓冲 channel、`close`、`for range`、`select`、`time.After`。

---

## 今天最容易踩的坑

### 坑 1：同一个 goroutine 中向无缓冲 channel 发送但没人接收

```go
ch := make(chan string)
ch <- "hello"
msg := <-ch
```

这会死锁，因为发送时已经阻塞，走不到接收那行。

需要另一个 goroutine 发送或接收。

---

### 坑 2：忘记 close 导致 for range 卡住

```go
for value := range ch {
    // ...
}
```

如果 channel 永远不关闭，这个循环可能一直等。

---

### 坑 3：接收方关闭 channel

通常由发送方关闭 channel，因为发送方知道什么时候不会再发送。

接收方关闭 channel 容易导致发送方继续发送时 panic。

---

### 坑 4：向已关闭 channel 发送

```go
close(ch)
ch <- "hello"
```

这会 panic。

---

### 坑 5：误以为 buffered channel 永远不阻塞

buffered channel 只有在缓冲区未满时发送不阻塞。

缓冲区满了仍然会阻塞。
