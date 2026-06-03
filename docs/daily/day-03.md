# Day 03 学习记录

## 日期

2026-06-03

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/3

## 今日教案

- 教案文档：`docs/daily/day-03-lesson.md`

## 核心任务

- 学习 Go 流程控制：`if/else`、`for`、`switch`、`break`、`continue`。
- 实现九九乘法表和基础分支判断练习。

## 验收标准

- 创建 `week01-basics/day03-control-flow/`。
- 至少实现 5 个流程控制函数。
- 至少写 5 组测试。
- `go test ./...` 通过。
- 能解释 `if/else`、`for`、`switch`、`break`、`continue` 和嵌套循环。

## 可选挑战题

- 实现一个简单命令行菜单。

## 学习前问题

- `nil` 是什么？

## 答疑记录

- `nil` 表示某些引用类型、指针类型或接口类型当前没有指向有效对象。
- `int`、`float64`、`string`、`bool` 等基础类型不能是 `nil`，它们有自己的零值。
- `slice`、`map`、`pointer`、`channel`、`function`、`interface` 可以是 `nil`。
- nil slice 可以 `append`，nil map 不能直接写入，nil pointer 不能直接解引用。

## 今日产出

- 保存 Day 3 教案：`docs/daily/day-03-lesson.md`。
- 创建 Day 3 练习目录：`week01-basics/day03-control-flow/`。
- 实现 5 个流程控制函数：`Grade`、`DayType`、`EvenNumbers`、`SumTo`、`MultiplicationTable`。
- 编写测试覆盖成绩分级、工作日判断、偶数筛选、累加和九九乘法表。

## 运行过的命令

```bash
go test ./week01-basics/day03-control-flow
make fmt
make test
make vet
```

## 代码 Review 结论

- `Grade` 使用 `if/else if/else` 完成范围判断，非法值优先判断，避免边界条件被误分类。
- `DayType` 使用 `switch` 匹配固定字符串，适合菜单、枚举、固定选项类逻辑。
- `EvenNumbers` 使用 `for`、`continue` 和 `append` 完成筛选，体现了 nil slice 可以安全 append。
- `SumTo` 使用累加器变量和 `for` 完成求和。
- `MultiplicationTable` 使用嵌套循环生成行和列，并用 `strings.Join` 控制列之间的空格。
- `make fmt`、`make test`、`make vet` 均已通过。

## 今日小测试

1. `if/else` 和 `switch` 分别适合什么场景？
   - 回答：分别适用范围判断和固定值匹配。
   - 结果：通过。
2. Go 里为什么说 `for` 是唯一的循环关键字？它可以写出哪几种循环形式？
   - 回答：可以写计数循环、条件循环、无限循环。
   - 结果：通过。
3. `break` 和 `continue` 的区别是什么？
   - 回答：`break` 退出当前循环；`continue` 跳过本轮循环，进入下一轮。
   - 结果：通过。
4. Go 的 `switch` 为什么通常不需要写 `break`？
   - 回答：默认不会继续执行下一个 case。
   - 结果：通过。
5. 在九九乘法表里，外层循环和内层循环分别负责什么？
   - 回答：外层循环控制行，内层循环控制每行的列。
   - 结果：通过。

## 测试结果

- 5 题全部通过。

## 遇到的问题

- `Grade` 中非法分数的返回值需要和测试约定保持一致：`Invalid`。
- `DayType` 中输入值和返回值需要严格匹配测试约定，例如 `Mon`、`Sat`、`weekday`、`weekend`。
- `EvenNumbers` 的循环边界一开始从 0 到 `< n`，导致结果包含 0 且漏掉 n；正确范围是从 1 到 `n`。
- `MultiplicationTable` 中 `fmt.Sprintf` 末尾多余空格会导致字符串测试失败，应交给 `strings.Join` 统一处理分隔符。

## 关键收获

1. `if/else` 适合范围判断，条件顺序很重要，非法值或特殊边界通常应优先判断。
2. `switch` 适合固定值匹配，Go 的 `switch` 默认不会贯穿到下一个 case，通常不需要手写 `break`。
3. `for` 是 Go 唯一的循环关键字，可以写计数循环、条件循环和无限循环。
4. `continue` 表示跳过当前这一轮，`break` 表示结束当前循环。
5. 嵌套循环适合生成二维结构，例如九九乘法表；外层控制行，内层控制列。

## 明日计划

- 进入 Day 4：函数。
- 练习函数定义、返回值、变参、匿名函数。
- 实现一个计算器练习。
