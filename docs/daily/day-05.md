# Day 05 学习记录

## 日期

2026-06-09

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/5

## 今日教案

- 教案文档：`docs/daily/day-05-lesson.md`

## 核心任务

- 学习 Go 错误处理：`error`、显式错误返回、错误包装、`defer`、`panic/recover`。
- 给计算器逻辑增加更明确的错误处理模式。

## 验收标准

- 创建 `week01-basics/day05-errors/`。
- 实现 `ErrDivideByZero`、`Divide`、`Calculate`、`IsDivideByZero`、`DeferOrder`、`SafeCall`。
- 至少写 6 组测试。
- `go test ./...` 通过。
- 能解释 `err == nil`、`errors.New`、`fmt.Errorf`、`%w`、`defer`、`panic` 和 `recover`。

## 可选挑战题

- 对计算器增加 typed errors 或更明确的错误分类。

## 学习前问题

- `defer` 是不是在函数返回之后才执行？
- 日常业务错误不应该优先使用 `panic`，那什么时候该用 `panic`？
- `recover` 是怎么捕获 `panic` 的？
- 具名返回值为什么能被 `defer` 修改？
- 有 `panic` 的情况下，没有执行普通 `return`，具名返回值还能返回吗？
- 如果不是具名返回值，能不能直接在 `defer` 里面 `return`？

## 答疑记录

- `defer` 不是在函数已经返回给调用方之后执行，而是在函数准备返回时执行；整体顺序可以理解为：设置返回值、执行所有 `defer`、真正返回给调用方。
- 普通业务错误应该返回 `error`，不应该用 `panic`；`panic` 更适合不可恢复的严重错误、框架边界兜底、服务入口保护、测试某段代码是否会崩溃等场景。
- `recover` 只有在 `defer` 中调用才有意义。发生 `panic` 后，Go 会沿调用栈执行已注册的 `defer`，如果某个 `defer` 中调用 `recover()`，就能拿到 panic 的值并停止继续崩溃。
- 具名返回值在函数开始执行时就已经存在，`return false` 会先把返回值设置为 `false`，随后执行 `defer`；如果 `defer` 闭包中修改了这个具名返回变量，最终返回的就是修改后的值。
- 发生 `panic` 时普通执行流程会中断，但已注册的 `defer` 仍会执行；如果 `defer` 中 `recover` 成功，函数可以继续完成返回，并返回当前具名返回值。
- 如果函数没有具名返回值，不能通过 `defer` 里的 `return` 直接替外层函数返回值；`defer` 中的 `return` 只会返回这个匿名函数本身。要让 `defer` 修改外层返回值，通常需要使用具名返回值。

## 今日产出

- 创建 Day 5 练习目录：`week01-basics/day05-errors/`。
- 实现错误处理练习函数：`ErrDivideByZero`、`ErrUnsupportedOperation`、`Divide`、`Calculate`、`IsDivideByZero`、`DeferOrder`、`SafeCall`。
- 编写测试覆盖正常除法、除零错误、计算器操作、错误包装、非法操作符、`defer` 执行顺序和 `panic/recover`。

## 运行过的命令

```bash
go test ./week01-basics/day05-errors
make fmt
make test
make vet
```

## 代码 Review 结论

- `Divide` 使用 `ErrDivideByZero` 直接表达除零这一原始错误，职责清晰。
- `Calculate` 在除法失败时使用 `fmt.Errorf` 和 `%w` 包装错误，为调用方提供更完整的上下文，同时保留原始错误链。
- `IsDivideByZero` 使用 `errors.Is`，能识别被 `%w` 包装后的除零错误。
- `DeferOrder` 使用具名返回值，让 `defer` 中的闭包可以修改最终返回的 slice，并正确体现后进先出的执行顺序。
- `SafeCall` 正确把 `recover` 放在 `defer` 中，用具名返回值表达是否发生过 panic。
- `go test ./week01-basics/day05-errors`、`make fmt`、`make test`、`make vet` 均已通过。

## 今日小测试

1. Go 中 `err == nil` 和 `err != nil` 分别表示什么？为什么调用函数后通常要先判断 `err`？
   - 回答：分别表示有错误和没有错误；避免错误影响外部程序正常运行。
   - 结果：需要修正。
   - 标准答案：`err == nil` 表示没有错误，函数调用成功；`err != nil` 表示发生错误。调用函数后通常要先判断 `err`，是为了先处理失败路径，避免继续使用无效结果。
2. `errors.New("divide by zero")` 创建的错误适合什么场景？为什么把它定义成 `var ErrDivideByZero` 有好处？
   - 回答：固定的、可枚举的错误状态；让调用方可以精确判断发生了什么错误。
   - 结果：通过。
   - 标准答案：`errors.New` 适合创建简单、固定、可复用的错误。定义成 `var ErrDivideByZero` 后，它可以作为 sentinel error 被复用，调用方可以用 `errors.Is` 精确判断错误类型。
3. `fmt.Errorf("calculate failed: %w", err)` 中 `%w` 和 `%v` 的区别是什么？
   - 回答：区别是能否追踪错误链，前者可以，后者不行。
   - 结果：通过。
   - 标准答案：`%w` 会包装错误并保留错误链，后续可以用 `errors.Is` 判断原始错误；`%v` 只把错误格式化成文本，不保留错误链。
4. 多个 `defer` 的执行顺序是什么？今天的 `DeferOrder` 为什么需要具名返回值？
   - 回答：后进先出；为了将更改后的值正确返回。
   - 结果：通过。
   - 标准答案：多个 `defer` 按后进先出执行。`DeferOrder` 需要具名返回值，是因为 `return` 会先设置返回值，再执行 `defer`；具名返回值可以在 `defer` 中被修改并影响最终返回结果。
5. `recover` 为什么必须放在 `defer` 中？普通业务错误为什么不应该优先用 `panic/recover` 处理？
   - 回答：只有放在 `defer` 中才能捕获 panic；普通业务错误更适合常规 errors，可预测可控。
   - 结果：通过。
   - 标准答案：`recover` 只有在 `defer` 中调用才能捕获正在传播的 `panic`。普通业务错误是可预期失败，应该通过 `error` 显式返回，让调用方正常处理，而不是用 `panic/recover` 改变普通控制流。

## 测试结果

- 4 题通过，1 题需要修正。
- 需要修正：`err == nil` 表示没有错误，`err != nil` 表示有错误。

## 遇到的问题

- 初次实现 `DeferOrder` 时，只修改了局部变量 `order`，没有影响已经准备返回的匿名返回值；改为具名返回值后，`defer` 可以修改最终返回值。
- 初次调整错误包装边界时，`Calculate` 仍然直接返回了 `ErrDivideByZero`；最终修正为 `Divide` 返回原始错误，`Calculate` 负责包装上下文。

## 关键收获

1. `error` 是显式错误返回机制，`nil` 表示成功，非 `nil` 表示失败，调用方需要主动处理错误。
2. sentinel error 适合表达可判断的固定错误，例如 `ErrDivideByZero`，调用方可以通过 `errors.Is` 判断。
3. `fmt.Errorf` 搭配 `%w` 可以在增加上下文的同时保留错误链，避免只剩错误文本。
4. 多个 `defer` 按后进先出执行；函数返回时会先设置返回值，再执行 `defer`，最后真正返回。
5. `panic/recover` 适合边界兜底或严重异常，不适合替代普通业务错误处理。

## 明日计划

- 进入 Day 6：包管理。
- 练习 package、import、init 和 go module。
- 组织 Week 1 的代码结构。
