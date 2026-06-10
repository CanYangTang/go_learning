# Day 05 教案：Go 错误处理

## 学习目标

学完今天，需要能够做到：

1. 理解 Go 为什么使用显式错误返回。
2. 使用 `error` 表示失败情况。
3. 使用 `errors.New` 创建简单错误。
4. 使用 `fmt.Errorf` 和 `%w` 包装错误。
5. 使用 `errors.Is` 判断被包装的错误。
6. 使用 `defer` 延迟执行清理逻辑。
7. 理解 `panic` 和 `recover` 的用途和限制。
8. 给 Day 4 的计算器逻辑增加更明确的错误处理模式。

---

## Go 的错误处理哲学

Go 不使用异常作为主要错误处理方式，而是让函数显式返回错误。

常见模式：

```go
result, err := doSomething()
if err != nil {
    return err
}
```

这个模式看起来啰嗦，但优点是：

- 错误路径非常明确。
- 调用者必须正视失败情况。
- 代码执行流程容易追踪。
- 不会在很深的调用栈里突然抛异常打断流程。

---

## error 是什么

`error` 是 Go 标准库中的一个接口：

```go
type error interface {
    Error() string
}
```

只要一个类型实现了 `Error() string` 方法，它就是一个 error。

日常最常见的是使用标准库创建错误：

```go
import "errors"

var ErrDivideByZero = errors.New("divide by zero")
```

---

## errors.New

`errors.New` 用来创建一个简单错误。

```go
import "errors"

func Divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("divide by zero")
    }
    return a / b, nil
}
```

调用：

```go
result, err := Divide(10, 0)
if err != nil {
    fmt.Println(err)
}
```

这里 `nil` 表示没有错误。

---

## nil error

Go 中常见约定：

```text
err == nil 表示成功
err != nil 表示失败
```

例如：

```go
result, err := Divide(10, 2)
if err != nil {
    return err
}
fmt.Println(result)
```

注意：

```go
nil
```

在这里表示没有错误。

---

## 从 bool 到 error

Day 4 的除法函数是：

```go
func Divide(a, b float64) (float64, bool)
```

今天会升级为：

```go
func Divide(a, b float64) (float64, error)
```

区别：

| 返回方式 | 表达能力 |
|---|---|
| `bool` | 只能表达成功或失败 |
| `error` | 能表达失败，并说明失败原因 |

例如：

```go
Divide(10, 0) // 0, ErrDivideByZero
```

这比 `0, false` 更清楚。

---

## sentinel error

sentinel error 是一个预定义的固定错误值。

```go
var ErrDivideByZero = errors.New("divide by zero")
```

它适合表达明确、可判断的错误类型。

调用方可以判断：

```go
if errors.Is(err, ErrDivideByZero) {
    fmt.Println("cannot divide by zero")
}
```

---

## fmt.Errorf

`fmt.Errorf` 可以格式化错误信息。

```go
err := fmt.Errorf("invalid operator: %s", op)
```

这适合错误信息里需要带变量的场景。

---

## 错误包装 %w

`fmt.Errorf` 搭配 `%w` 可以包装已有错误。

```go
err := fmt.Errorf("calculate failed: %w", ErrDivideByZero)
```

这样既能保留上下文：

```text
calculate failed: divide by zero
```

又能保留原始错误，让调用者判断：

```go
errors.Is(err, ErrDivideByZero)
```

---

## errors.Is

`errors.Is` 用来判断一个错误链中是否包含指定错误。

```go
err := fmt.Errorf("calculate failed: %w", ErrDivideByZero)

if errors.Is(err, ErrDivideByZero) {
    fmt.Println("matched divide by zero")
}
```

如果错误被 `%w` 包装过，`errors.Is` 仍然能识别。

如果只是 `%v`：

```go
fmt.Errorf("calculate failed: %v", ErrDivideByZero)
```

这只是把错误转成文本，不会保留错误链。

---

## defer

`defer` 用来延迟执行一段代码，通常用于清理资源。

```go
func ReadFile(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()

    // read file
    return nil
}
```

`defer file.Close()` 会在当前函数返回前执行。

常见用途：

- 关闭文件
- 释放锁
- 关闭连接
- 打印函数退出日志
- 配合 recover 捕获 panic

---

## defer 执行顺序

如果有多个 defer，它们按后进先出执行。

```go
func Demo() {
    defer fmt.Println("first")
    defer fmt.Println("second")
    defer fmt.Println("third")
}
```

输出：

```text
third
second
first
```

记忆方式：像栈一样，后放进去的先执行。

---

## defer 的参数会立即求值

```go
func Demo() {
    name := "Alice"
    defer fmt.Println(name)
    name = "Bob"
}
```

输出：

```text
Alice
```

因为 `defer fmt.Println(name)` 注册时，参数 `name` 已经被求值。

---

## panic

`panic` 表示程序遇到无法正常处理的严重问题。

```go
panic("something went wrong")
```

触发 panic 后，当前函数会停止正常执行，开始向外层调用栈传播。

日常业务错误不应该优先使用 panic。

不推荐：

```go
func Divide(a, b float64) float64 {
    if b == 0 {
        panic("divide by zero")
    }
    return a / b
}
```

推荐：

```go
func Divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, ErrDivideByZero
    }
    return a / b, nil
}
```

---

## recover

`recover` 可以在 `defer` 中捕获 panic，避免程序继续崩溃。

```go
func SafeCall(fn func()) (panicked bool) {
    defer func() {
        if r := recover(); r != nil {
            panicked = true
        }
    }()

    fn()
    return false
}
```

注意：

- `recover` 只有在 `defer` 中调用才有意义。
- `recover` 只能捕获当前 goroutine 中传播的 panic。
- 不要用 panic/recover 代替普通错误处理。

---

## panic/recover 适合什么场景

适合：

- 框架边界兜底
- 服务入口保护
- 测试某段代码是否 panic
- 不可恢复的严重错误

不适合：

- 普通参数校验
- 数据库查询失败
- HTTP 请求失败
- 用户输入错误
- 除数为 0 这种可预期错误

可预期错误应该用 `error`。

---

## 今日代码练习设计

今天会写一个小 package：

```text
week01-basics/day05-errors/
├── errors.go
└── errors_test.go
```

建议实现这些内容：

### 定义除零错误

```go
var ErrDivideByZero = errors.New("divide by zero")
```

练习点：

- sentinel error
- `errors.New`

### 带 error 的除法

```go
func Divide(a, b float64) (float64, error)
```

期望：

```go
Divide(10, 2) // 5, nil
Divide(10, 0) // 0, ErrDivideByZero
```

练习点：

- 显式错误返回
- `err == nil` 表示成功

### 计算器操作

```go
func Calculate(a, b float64, op string) (float64, error)
```

规则：

```text
+ -> 加法
- -> 减法
* -> 乘法
/ -> 除法
其他 -> invalid operator 错误
```

练习点：

- `switch`
- 错误返回
- 错误包装

### 判断错误是否为除零

```go
func IsDivideByZero(err error) bool
```

练习点：

- `errors.Is`

### defer 顺序

```go
func DeferOrder() []string
```

期望返回：

```go
[]string{"third", "second", "first"}
```

练习点：

- 多个 defer 的执行顺序

### 安全调用

```go
func SafeCall(fn func()) (panicked bool)
```

期望：

```go
SafeCall(func() {}) // false
SafeCall(func() { panic("boom") }) // true
```

练习点：

- `panic`
- `recover`
- `defer`
- 具名返回值

---

## 今天最容易踩的坑

### 坑 1：忘记检查 err

错误：

```go
result, _ := Divide(10, 0)
fmt.Println(result)
```

更好的写法：

```go
result, err := Divide(10, 0)
if err != nil {
    return err
}
fmt.Println(result)
```

### 坑 2：用 panic 处理普通错误

普通业务错误应该返回 `error`，不是 panic。

### 坑 3：用 `%v` 导致错误链丢失

```go
fmt.Errorf("calculate failed: %v", ErrDivideByZero)
```

这不能被 `errors.Is` 识别。

应该用：

```go
fmt.Errorf("calculate failed: %w", ErrDivideByZero)
```

### 坑 4：recover 不在 defer 中使用

```go
recover()
```

这样无法捕获 panic。

应该写在 defer 中：

```go
defer func() {
    if r := recover(); r != nil {
        // handle panic
    }
}()
```

### 坑 5：误解 defer 参数求值时机

`defer` 注册时参数会立即求值，不是等到真正执行时才求值。

---

## 今日验收标准

今天完成后应该满足：

1. 创建 `week01-basics/day05-errors/`。
2. 实现 `ErrDivideByZero`、`Divide`、`Calculate`、`IsDivideByZero`、`DeferOrder`、`SafeCall`。
3. 至少写 6 组测试。
4. `go test ./...` 通过。
5. 能解释：
   - `err == nil` 和 `err != nil` 的含义
   - `errors.New` 和 `fmt.Errorf` 的区别
   - `%w` 为什么能保留错误链
   - `defer` 的执行顺序
   - 为什么普通业务错误不应该用 panic
   - `recover` 为什么要放在 defer 中
6. 更新 `docs/daily/day-05.md`。
