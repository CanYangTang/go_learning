# Day 06 学习记录

## 日期

2026-06-10

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/6

## 今日教案

- 教案文档：`docs/daily/day-06-lesson.md`

## 核心任务

- 学习 Go 包管理：`package`、`import`、导出标识符、`go.mod`、module path、import path、`init`。
- 拆分练习代码到多个 package，并练习项目内部 package 的导入。

## 验收标准

- 创建 `week01-basics/day06-packages/`。
- 创建或补充 `pkg/mathutil/`。
- 实现 `Double`、`IsEven`、`NormalizeName`、`UserLabel`、`DoubleAge`、`IsAdultAgeEven`、`PackageStatus`。
- 至少写 7 组测试。
- `go list ./...` 能看到新增 package。
- `go test ./...` 通过。
- 能解释 package 和目录的关系、module path 和 import path 的关系、导出标识符为什么要大写、`init` 什么时候执行。

## 可选挑战题

- 创建一个可复用工具 package，并补充测试。

## 学习前问题

-

## 答疑记录

-

## 今日产出

- 创建可复用工具 package：`pkg/mathutil/`。
- 创建 Day 6 练习 package：`week01-basics/day06-packages/`。
- 实现 `Double`、`IsEven`、`NormalizeName`、`UserLabel`、`DoubleAge`、`IsAdultAgeEven`、`PackageStatus`。
- 编写测试覆盖工具函数、字符串格式化、项目内部 package 导入、`init` 初始化状态。

## 运行过的命令

```bash
go test ./pkg/mathutil
go test ./week01-basics/day06-packages
go list ./...
make fmt
make test
make vet
```

## 代码 Review 结论

- `pkg/mathutil` 作为可复用工具 package，函数名使用大写开头，能被其他 package 导入调用。
- `day06-packages` 正确导入标准库 `fmt`、`strings` 和项目内部 package `github.com/CanYangTang/go_learning/pkg/mathutil`。
- `NormalizeName` 使用 `strings.TrimSpace`，比只裁剪普通空格更通用。
- `UserLabel` 复用 `NormalizeName`，避免调用方传入带空格名字时格式异常。
- `IsAdultAgeEven` 同时判断 `age >= 18` 和偶数，符合函数语义。
- `PackageStatus` 通过 package 级变量和 `init` 初始化，能体现 `init` 的自动执行时机。
- `go list ./...` 能看到新增 package，`make fmt`、`make test`、`make vet` 均已通过。

## 今日小测试

1. Go 中 package 和目录是什么关系？同一个目录里通常能不能混用多个普通 package？
   - 回答：强绑定关系；不能。
   - 结果：通过。
   - 标准答案：Go 以 package 为编译单位，一个目录通常对应一个 package；同一个目录里通常不能混用多个普通 package，否则会编译失败。
2. `go.mod` 里的 module path 和代码里的 import path 有什么关系？以今天的 `pkg/mathutil` 为例说明。
   - 回答：代表项目的根路径，`github.com/CanYangTang/go_learning/pkg/mathutil`。
   - 结果：通过。
   - 标准答案：`go.mod` 里的 module path 表示项目根导入路径；项目内部 package 的 import path 通常是 module path 加目录路径，例如 `github.com/CanYangTang/go_learning/pkg/mathutil`。
3. 为什么 `Double` 和 `IsEven` 要首字母大写？如果写成 `double`，其他 package 能调用吗？
   - 回答：为了能被外部包使用；不能。
   - 结果：通过。
   - 标准答案：Go 使用首字母大小写控制可见性，首字母大写的标识符可以被其他 package 调用；如果写成 `double`，它只在当前 package 内可见，其他 package 不能调用。
4. `import "github.com/CanYangTang/go_learning/pkg/mathutil"` 导入的是文件路径还是目录路径？调用时为什么写 `mathutil.Double(3)`？
   - 回答：目录路径；调用时需要使用 package 名。
   - 结果：通过。
   - 标准答案：`import` 导入的是 package 所在的目录路径，不是某个 `.go` 文件；调用时使用源码中声明的 package 名，所以写 `mathutil.Double(3)`。
5. `init` 函数什么时候执行？它适合做什么，不适合做什么？
   - 回答：初始化导入的 package 和当前 package 的变量后执行；适合注册驱动、插件、初始化固定映射表、做非常轻量的 package 初始化。不适合启动复杂业务逻辑、发 HTTP 请求、连接数据库、依赖外部环境的重操作等。
   - 结果：通过。
   - 标准答案：`init` 在 package 初始化时自动执行，通常在依赖 package 初始化完成、本 package 变量初始化后执行；它适合注册驱动、插件、初始化固定映射表等轻量初始化，不适合隐藏复杂业务逻辑或依赖外部环境的重操作。

## 测试结果

- 5 题全部通过。

## 遇到的问题

- 初次实现 `UserLabel` 时没有复用 `NormalizeName`，导致带空格的名字被直接格式化进结果。
- 初次实现 `IsAdultAgeEven` 时只判断了偶数，没有同时判断是否成年。

## 关键收获

1. Go 以 package 为编译和测试单位，一个目录通常对应一个 package。
2. 项目内部 import path 通常由 `go.mod` 中的 module path 加目录路径组成。
3. 首字母大写的标识符可以被其他 package 使用，首字母小写的标识符只在当前 package 内可见。
4. `go list ./...` 可以检查当前 module 下有哪些 package，以及新增目录是否被 Go 识别。
5. `init` 会在 package 初始化时自动执行，适合轻量初始化，不适合隐藏复杂业务逻辑。

## 明日计划

- 进入 Day 7：Week 1 总结与工具包练习。
- 复盘 Week 1 的语法、控制流、函数、错误处理和包管理。
- 整理 Week 1 的代码结构和学习记录。
