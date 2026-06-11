# Day 07 学习记录

## 日期

2026-06-11

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/7

## 今日教案

- 教案文档：`docs/daily/day-07-lesson.md`

## 核心任务

- 完成一个小型可复用 utility package。
- 编写 Week 1 学习复盘，总结 Go 基础、测试、错误处理和包管理。

## 验收标准

- 创建 `pkg/stringutil/`。
- 实现 `IsBlank`、`NormalizeSpace`、`JoinNonEmpty`。
- 为 `pkg/stringutil` 编写测试。
- 更新 `docs/weekly/week-01.md`。
- `make fmt`、`make test`、`make vet` 通过。
- 能解释 Week 1 的核心知识地图和仍需巩固的点。

## 可选挑战题

- 为 utility package 增加使用说明或更多边界测试。

## 学习前问题

- 前面的学习只覆盖了基础类型，还没有系统学习 slice，但 Day 7 的工具函数会用到 slice。

## 答疑记录

- Day 7 教案已补充 slice 基础专题，覆盖 slice 的声明、`append`、`len`、`for range`、变参和 slice 的关系，以及它在 `JoinNonEmpty` 中的使用方式。
- `go.mod` 是 Go 项目的模块配置文件，主要定义 module path、Go 版本和第三方依赖；项目内部 import path 通常基于 `go.mod` 中的 module path。
- `Makefile` 是常用开发命令的快捷入口，把 `go fmt ./...`、`go test ./...`、`go vet ./...` 等命令封装成 `make fmt`、`make test`、`make vet`，便于重复执行。

## 今日产出

- 创建 `pkg/stringutil/` 工具 package。
- 实现 `IsBlank`、`NormalizeSpace`、`JoinNonEmpty`。
- 编写测试覆盖空字符串、空白字符串、连续空白、变参为空和部分参数为空等边界。
- 更新 Week 1 周总结：`docs/weekly/week-01.md`。

## 运行过的命令

```bash
go test ./pkg/stringutil
go list ./...
make fmt
make test
make vet
```

## 代码 Review 结论

- `IsBlank` 使用 `strings.TrimSpace` 判断空白字符串，能处理普通空格、tab 和换行。
- `NormalizeSpace` 使用 `strings.Fields` 和 `strings.Join`，能同时完成裁剪两端空白和压缩中间连续空白。
- `JoinNonEmpty` 先标准化每个 part，再过滤空字符串，最后用指定分隔符拼接，逻辑顺序清晰。
- `JoinNonEmpty` 使用变参和 slice，能体现 `parts ...string` 在函数内部作为 `[]string` 使用的方式。
- 测试覆盖了空输入、全空白、连续空白、变参展开和无参数场景。
- `go list ./...` 能看到新增 package，`make fmt`、`make test`、`make vet` 均已通过。

## 今日小测试

1. slice 和数组有什么区别？为什么日常 Go 代码中更常用 slice？
   - 回答：区别是 slice 的长度是动态的，数组是固定的；因为更方便。
   - 结果：通过。
   - 标准答案：数组长度固定，长度也是类型的一部分；slice 是基于数组的动态视图，长度可以变化，能配合 `append` 使用，因此日常 Go 代码中更常用 slice。
2. 为什么 `append` 的返回值通常要重新赋给原 slice？
   - 回答：这样原 slice 才能同步到最新的值。
   - 结果：通过。
   - 标准答案：`append` 可能返回一个新的 slice header，甚至可能因为容量不足而分配新的底层数组；如果不接住返回值，调用方看到的 slice 不会更新。
3. `parts ...string` 在函数内部可以当作什么类型使用？如果已有 `[]string`，调用变参函数时要怎么传入？
   - 回答：slice；`parts...`。
   - 结果：通过。
   - 标准答案：`parts ...string` 在函数内部可以当作 `[]string` 使用；如果已有 `items := []string{"a", "b"}`，调用变参函数时使用 `JoinNonEmpty(",", items...)` 展开传入。
4. `strings.TrimSpace` 和 `strings.Fields` 分别适合解决什么问题？今天的 `NormalizeSpace` 为什么可以用 `Fields` + `Join` 实现？
   - 回答：分别是将字符串的首尾空白字符去除和忽略首尾空白字符并将字符按连续空白字符分隔成序列。
   - 结果：通过。
   - 标准答案：`strings.TrimSpace` 适合去掉字符串首尾空白；`strings.Fields` 会按连续空白切分字符串并忽略首尾空白。`NormalizeSpace` 可以先用 `Fields` 得到非空字段，再用 `strings.Join(fields, " ")` 用单个空格拼回去。
5. 本周你认为最容易混淆的三个概念是什么？分别应该如何区分？
   - 回答：如 `defer` 修改局部变量和具名返回值的区别；`string` 和 `strconv.Itoa` 等。
   - 结果：部分通过。
   - 标准答案：可以重点区分三组概念：第一，`defer` 修改局部变量不一定影响已设置好的匿名返回值，但修改具名返回值会影响最终返回；第二，`string(123)` 是把整数当作 Unicode 码点转换，`strconv.Itoa(123)` 才是把数字转成字符串 `"123"`；第三，函数作为参数、匿名函数和闭包不同，函数作为参数只要求签名匹配，匿名函数是没有名字的函数，闭包是函数捕获并持续引用外部变量。

## 测试结果

- 4 题通过，1 题部分通过。
- 需要补充：复盘题要尽量列满三个易混概念，并写清楚每组概念的区别。

## 遇到的问题

- Day 7 开始前发现前面只系统学习了基础类型，还没有系统学习 slice，因此补充了 slice 基础专题。
- 初次实现 `JoinNonEmpty` 时，判断使用了标准化后的字符串，但 append 的仍是原始字符串，导致结果保留了多余空白。

## 关键收获

1. slice 是 Go 中常用的动态长度序列，可以配合 `append`、`len` 和 `for range` 使用。
2. 变参 `parts ...string` 在函数内部可以当作 `[]string` 使用，已有 slice 调用变参函数时可以用 `items...` 展开。
3. `strings.Fields` 可以按连续空白切分字符串，再用 `strings.Join` 拼接，可以实现空白标准化。
4. 工具函数应该保持输入输出明确、逻辑独立、容易测试，不要为了工程化而过早抽象。
5. 周复盘要记录掌握点、薄弱点和典型错误，比单纯记录“学了什么”更有价值。

## 明日计划

- 进入 Week 2 Day 8：goroutine 并发基础。
- 学习并发执行、`sync.WaitGroup` 和基础并发测试。
