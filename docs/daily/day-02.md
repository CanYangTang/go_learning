# Day 02 学习记录

## 日期

2026-06-03

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/2

## 今日教案

- 教案文档：`docs/daily/day-02-lesson.md`

## 核心任务

- 练习变量、常量、基础类型和类型转换。
- 编写可运行的 Go 示例和测试。

## 验收标准

- 创建 `week01-basics/day02-syntax/`。
- 至少实现 5 个基础函数。
- 至少写 5 个测试。
- `go test ./...` 通过。
- 能解释 `var`、`:=`、`const`、零值和显式类型转换。

## 可选挑战题

- 总结 `int`、`float64`、`string` 转换的常见坑。

## 学习前问题

- `fmt.Sprintf` 和 `fmt.Printf` 的区别是什么？
- `int64` 和 `int` 的区别是什么？

## 答疑记录

- `fmt.Sprintf` 会格式化并返回字符串，适合生成返回值、日志内容或测试期望值；`fmt.Printf` 会格式化并直接打印到终端，适合命令行输出或临时调试。
- `int` 的大小和运行平台有关，日常数量、下标、年龄等普通整数优先用 `int`；`int64` 固定是 64 位，适合数据库大 ID、时间戳和外部接口明确要求 64 位整数的场景。

## 今日产出

- 保存 Day 2 教案：`docs/daily/day-02-lesson.md`。
- 更新每日 AI 工作流和 daily 模板。
- 创建 Day 2 练习目录：`week01-basics/day02-syntax/`。
- 实现 5 个基础函数：`UserProfile`、`RectangleArea`、`Average`、`IsAdult`、`FormatScore`。
- 编写 5 组测试，覆盖字符串格式化、浮点计算、显式类型转换、布尔判断和数字转字符串。

## 运行过的命令

```bash
go test ./week01-basics/day02-syntax
make fmt
make test
make vet
```

## 代码 Review 结论

- `UserProfile` 使用 `fmt.Sprintf` 直接返回字符串，表达清晰。
- `Average` 正确使用 `float64(total) / float64(count)`，避免了整数除法。
- `FormatScore` 使用 `strconv.Itoa` 练习了 `int` 到 `string` 的正确转换。
- 所有测试和 `go vet` 均已通过。

## 今日小测试

1. `var age int = 18`、`var age = 18`、`age := 18` 三种写法有什么区别？
   - 回答：分别是显式声明、类型推断声明和短变量声明；短变量声明只能在函数内部使用。
   - 结果：通过。
2. 什么是 Go 的零值？请分别说出 `int`、`float64`、`string`、`bool` 的零值。
   - 回答：变量只声明不赋值时会获得零值，分别是 `0`、`0`、`""`、`false`。
   - 结果：通过。
3. 为什么 `int` 和 `float64` 直接相乘不能编译？
   - 回答：因为短变量声明只能在函数内部使用。
   - 结果：需要修正。
   - 补充：真正原因是 `total` 是 `int`，`rate` 是 `float64`，Go 不会做隐式类型转换，必须写成 `float64(total) * rate`。`:=` 只要在函数内部使用就是合法的。
4. `string(123)` 和 `strconv.Itoa(123)` 的区别是什么？
   - 回答：`string(123)` 会把 123 当作 Unicode 码点得到对应字符，而 `strconv.Itoa(123)` 会转换为数字 123。
   - 结果：基本通过。
   - 补充：`strconv.Itoa(123)` 返回的是字符串 `"123"`，不是数字 123。
5. 为什么 `Average` 要写成 `float64(total) / float64(count)`，而不是 `total / count`？
   - 回答：需要避免整数除法，否则得到的值会是整数。
   - 结果：通过。

## 测试结果

- 4 题通过，1 题需要修正。
- 需要重点巩固：Go 不做隐式类型转换；`strconv.Itoa` 返回字符串。

## 遇到的问题

- 容易把 `:=` 的使用限制和类型不匹配编译错误混在一起。
- 容易把 `strconv.Itoa(123)` 的结果说成数字，实际结果是字符串。

## 关键收获

1. `var age int = 18` 是显式声明类型，`var age = 18` 是类型推断，`age := 18` 是函数内部使用的短变量声明。
2. Go 的零值让变量在只声明不赋值时也有确定状态：`int` 是 `0`，`float64` 是 `0`，`string` 是 `""`，`bool` 是 `false`。
3. Go 不做隐式类型转换，不同数值类型参与运算时必须显式转换。
4. `string(123)` 不是数字转字符串，而是把 123 当作 Unicode 码点；数字转字符串应使用 `strconv.Itoa(123)`。
5. 整数除法会丢掉小数部分，计算平均值时要先转换为 `float64`。

## 明日计划

- 进入 Day 3：流程控制。
- 练习 `if/else`、`for`、`switch`。
- 实现九九乘法表和基础分支判断练习。
