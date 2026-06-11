# Week 01 复盘：Go Foundations

## 时间范围

2026-06-01 ~ 2026-06-11

## 本周目标

- 建立 Go 学习项目的基础工程结构。
- 掌握 Go 基础语法、控制流、函数、错误处理和包管理。
- 通过每日代码练习、测试、Review 和小测试形成完整学习闭环。

## 已完成 Issue

- Day 1: Initialize Go learning backend project
- Day 2: Implement Go variables/types exercises
- Day 3: Implement control-flow exercises
- Day 4: Implement function and calculator exercises
- Day 5: Implement error handling examples
- Day 6: Organize packages and module structure
- Day 7: Build utility package and weekly review

## 本周代码产出

- 初始化 Go module、基础目录结构、Makefile 和最小 HTTP 服务骨架。
- 完成 `week01-basics/day02-syntax/`，练习变量、基础类型、类型转换和字符串格式化。
- 完成 `week01-basics/day03-control-flow/`，练习 `if`、`for`、`switch`、`break`、`continue` 和嵌套循环。
- 完成 `week01-basics/day04-functions/`，实现计算器函数、变参函数、函数作为参数和闭包计数器。
- 完成 `week01-basics/day05-errors/`，实现 sentinel error、错误包装、`errors.Is`、`defer`、`panic/recover`。
- 完成 `week01-basics/day06-packages/` 和 `pkg/mathutil/`，练习 package、import、导出标识符和 `init`。
- 完成 `pkg/stringutil/`，练习 slice、变参、字符串工具函数和边界测试。

## 掌握的概念

- `go.mod` 定义 module path，项目内部 package 的 import path 通常基于该路径。
- Makefile 可以封装常用开发命令，例如 `make fmt`、`make test`、`make vet`。
- Go 变量有显式声明、类型推断声明和短变量声明，未赋值变量会获得零值。
- Go 不做隐式类型转换，整数除法和字符串转换需要特别注意。
- `if`、`for`、`switch` 是 Go 中最常用的控制流结构。
- 函数可以返回多个值，变参在函数内部可以当作 slice 使用。
- 匿名函数可以作为参数传递，闭包可以捕获外部变量并保存状态。
- `error` 用于显式表达失败，`err == nil` 表示成功，`err != nil` 表示失败。
- `fmt.Errorf` 搭配 `%w` 可以包装错误并保留错误链，`errors.Is` 可以识别被包装的 sentinel error。
- 多个 `defer` 按后进先出执行，具名返回值可以被 `defer` 中的闭包修改。
- Go 以 package 为编译和测试单位，导出标识符需要首字母大写。
- slice 是动态长度序列，常和 `append`、`len`、`for range`、变参函数一起使用。

## 仍然薄弱的点

- `err == nil` 和 `err != nil` 的语义需要继续强化，避免把成功和失败反过来。
- 函数作为参数、匿名函数和闭包之间的边界需要继续通过代码练习区分。
- `defer`、return、具名返回值之间的执行顺序还需要通过更多例子巩固。
- slice 还只是补充入门，后续还需要系统学习底层数组、长度、容量、切片表达式和共享底层数组。
- package 拆分目前能完成基础导入，后续需要结合更真实的业务分层继续练习。

## 典型错误与修正

- Day 2 曾忘记为 `fmt.Sprintf` 和 `strconv.Itoa` 添加 import，修正后理解了 Go 对未使用/未导入 package 的严格检查。
- Day 2 曾混淆 `string(123)` 和 `strconv.Itoa(123)`，修正后理解前者是 Unicode 码点转换，后者是数字转字符串。
- Day 3 曾在 `Grade`、`DayType`、`EvenNumbers`、九九乘法表格式中出现规则偏差，修正后加强了测试驱动实现的意识。
- Day 4 曾把函数作为参数误认为闭包，把匿名函数作为参数误认为立即执行，修正后明确了三个概念的区别。
- Day 5 曾在 `DeferOrder` 中修改局部变量但没有影响返回值，修正后理解了具名返回值与 `defer` 的关系。
- Day 5 曾把错误包装放在底层 `Divide`，修正后改为底层返回原始错误、上层 `Calculate` 包装上下文。
- Day 6 曾在 `UserLabel` 中没有复用 `NormalizeName`，在 `IsAdultAgeEven` 中只判断偶数，修正后加强了函数语义和组合意识。
- Day 7 曾在 `JoinNonEmpty` 中判断了标准化结果但 append 原始字符串，修正后统一使用标准化后的值。

## 代码 Review 结论

- Week 1 的练习代码总体保持小函数、纯函数、测试驱动，适合基础阶段学习和复盘。
- 每天的测试都覆盖了核心行为和至少一个边界场景，能帮助及时暴露语义偏差。
- `pkg/mathutil` 和 `pkg/stringutil` 已开始形成可复用工具包，但当前仍应保持小而明确，避免过早抽象。
- 错误处理练习已经体现了 sentinel error、错误包装和错误判断的基本分层思路。
- 后续进入 HTTP、JSON、并发和 Gin/GORM 时，应继续保持“先定义行为，再写测试，再实现”的节奏。

## 验证命令

```bash
make fmt
make test
make vet
go list ./...
```

## 下周优先级

1. 学习 goroutine、channel 和基础并发模型。
2. 学习 HTTP server、JSON request/response 和文件操作。
3. 把并发与 HTTP 结合起来，完成 Week 2 综合练习。
