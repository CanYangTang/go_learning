# Day 06 教案：Go 包管理与模块结构

## 学习目标

学完今天，需要能够做到：

1. 理解 Go 中 `package` 的作用。
2. 理解一个目录、一个 package、多个 `.go` 文件之间的关系。
3. 使用 `import` 引入标准库和项目内部 package。
4. 理解导出标识符和非导出标识符。
5. 理解 `go.mod`、module path 和 import path 的关系。
6. 使用 `go list ./...` 查看项目中的 package。
7. 理解 `init` 函数的执行时机和适用场景。
8. 把 Week 1 的练习组织成可复用 package。

---

## package 是什么

Go 程序不是以单个文件为编译单位，而是以 package 为编译单位。

一个 package 通常对应一个目录：

```text
week01-basics/day06-packages/
├── calculator.go
├── format.go
└── calculator_test.go
```

如果这三个文件第一行都是：

```go
package packages
```

那么它们就属于同一个 package，可以直接互相使用对方定义的函数、变量和类型。

---

## 目录和 package 的关系

Go 的常见规则是：

```text
一个目录中只能有一个普通 package
```

例如：

```text
week01-basics/day06-packages/
```

目录下所有非测试 `.go` 文件通常都应该写同一个 package 名：

```go
package packages
```

不能在同一个目录里混着写：

```go
package packages
package utils
```

否则 Go 编译器会报错。

例外是测试文件可以使用：

```go
package packages
```

或：

```go
package packages_test
```

今天先使用和源码相同的 package 名，降低理解成本。

---

## package 名和目录名一定要相同吗

不强制相同，但推荐保持一致或语义接近。

例如目录：

```text
week01-basics/day06-packages/
```

因为目录名里有连字符 `-`，不能作为 package 名，所以源码里可以使用：

```go
package packages
```

Go package 名必须是合法标识符，不能包含 `-`。

---

## main package

`package main` 是特殊 package，表示这个 package 可以被编译成可执行程序。

例如：

```go
package main

func main() {
    // 程序入口
}
```

当前项目里的服务入口是：

```text
cmd/server/main.go
```

它使用 `package main`，因为它最终要通过：

```bash
go run ./cmd/server
```

启动程序。

普通练习目录不需要写 `package main`，因为它们更适合作为可测试、可复用的 package。

---

## import 是什么

`import` 用来引入其他 package。

标准库示例：

```go
import "fmt"

func FormatName(name string) string {
    return fmt.Sprintf("name=%s", name)
}
```

引入多个 package：

```go
import (
    "fmt"
    "strings"
)
```

Go 要求：导入了 package 就要使用，否则编译失败。

这能保持代码干净，避免无用依赖。

---

## 导出和非导出

Go 通过首字母大小写控制可见性。

首字母大写：可以被其他 package 使用。

```go
func Add(a, b int) int {
    return a + b
}
```

首字母小写：只能在当前 package 内部使用。

```go
func normalizeName(name string) string {
    return strings.TrimSpace(name)
}
```

如果另一个 package 想调用函数，函数名必须大写。

---

## 项目内部 import

当前项目的 `go.mod` 中定义了 module path：

```go
module github.com/CanYangTang/go_learning
```

这意味着项目内部 package 的完整 import path 通常从这个 module path 开始。

例如如果有目录：

```text
pkg/mathutil/
```

其中源码是：

```go
package mathutil

func Double(n int) int {
    return n * 2
}
```

其他 package 可以这样导入：

```go
import "github.com/CanYangTang/go_learning/pkg/mathutil"

func Demo() int {
    return mathutil.Double(3)
}
```

注意：import 用的是目录路径，不是 package 声明名。

但调用时使用的是 package 名：

```go
mathutil.Double(3)
```

---

## go.mod 的作用

`go.mod` 是 Go module 的配置文件。

它至少包含：

```go
module github.com/CanYangTang/go_learning

go 1.24.5
```

它说明：

- 当前项目的 module path 是什么。
- 当前项目使用的 Go 版本。
- 当前项目依赖哪些第三方 module。

今天不需要新增第三方依赖，但要理解：项目内部 import path 与 `module` 这一行直接相关。

---

## module path、import path、package name

这三个概念容易混淆。

### module path

定义在 `go.mod`：

```go
module github.com/CanYangTang/go_learning
```

表示整个项目的根路径。

### import path

用于导入某个 package：

```go
import "github.com/CanYangTang/go_learning/pkg/mathutil"
```

通常是：

```text
module path + 目录路径
```

### package name

写在源码第一行：

```go
package mathutil
```

调用时用 package name：

```go
mathutil.Double(3)
```

总结：

| 概念 | 位置 | 作用 |
|---|---|---|
| module path | `go.mod` | 定义项目根导入路径 |
| import path | `import` | 告诉 Go 去哪里找 package |
| package name | `.go` 文件第一行 | 代码中调用该 package 的名字 |

---

## go list ./...

`go list` 用来列出 package。

```bash
go list ./...
```

含义是：列出当前 module 下所有 package。

输出类似：

```text
github.com/CanYangTang/go_learning/cmd/server
github.com/CanYangTang/go_learning/pkg/response
github.com/CanYangTang/go_learning/week01-basics/day06-packages
```

它适合用来检查：

- 当前目录是否被 Go 识别为 package。
- package 路径是否符合预期。
- 是否有 package 编译结构错误。

---

## init 函数

`init` 是 Go 中特殊函数。

```go
func init() {
    // package 初始化时执行
}
```

特点：

- 不需要手动调用。
- 没有参数。
- 没有返回值。
- 在 package 被初始化时自动执行。
- 一个 package 里可以有多个 `init`。

执行顺序大致是：

1. 初始化被导入的 package。
2. 初始化当前 package 的变量。
3. 执行当前 package 的 `init`。
4. 如果是 `main` package，最后执行 `main`。

---

## init 适合做什么

适合：

- 注册驱动。
- 注册插件。
- 初始化固定映射表。
- 做非常轻量的 package 初始化。

不适合：

- 启动复杂业务逻辑。
- 发 HTTP 请求。
- 连接数据库。
- 依赖外部环境的重操作。
- 隐藏太多副作用。

学习阶段要记住：`init` 会自动执行，但不要滥用。

---

## 今日代码练习设计

今天会写两个小 package：

```text
week01-basics/day06-packages/
├── profile.go
└── profile_test.go

pkg/mathutil/
├── mathutil.go
└── mathutil_test.go
```

`pkg/mathutil` 作为可复用工具 package。

`week01-basics/day06-packages` 作为练习 package，导入 `pkg/mathutil`。

---

## 建议实现内容

### pkg/mathutil

实现：

```go
func Double(n int) int
func IsEven(n int) bool
```

期望：

```go
Double(3) // 6
IsEven(4) // true
IsEven(5) // false
```

练习点：

- 创建可复用 package。
- 函数首字母大写表示导出。
- package 能被其他目录导入。

---

## day06-packages

实现：

```go
func NormalizeName(name string) string
func UserLabel(name string, age int) string
func DoubleAge(age int) int
func IsAdultAgeEven(age int) bool
func PackageStatus() string
```

规则：

```text
NormalizeName: 去掉名字两边空格
UserLabel: 返回 "Alice(18)"
DoubleAge: 调用 mathutil.Double
IsAdultAgeEven: age >= 18 且 age 是偶数
PackageStatus: 返回一个由 init 设置的状态值
```

练习点：

- 导入标准库 `fmt`、`strings`。
- 导入项目内部 package `github.com/CanYangTang/go_learning/pkg/mathutil`。
- 使用导出函数。
- 使用 `init` 初始化 package 内部变量。

---

## 参考结构

```go
package packages

import (
    "fmt"
    "strings"

    "github.com/CanYangTang/go_learning/pkg/mathutil"
)

var status string

func init() {
    status = "ready"
}
```

这里：

- `status` 小写，只能在当前 package 内部使用。
- `PackageStatus` 可以大写导出，让测试读取状态。
- `mathutil.Double` 来自另一个 package。

---

## 今日验收标准

完成后应该满足：

1. 创建 `week01-basics/day06-packages/`。
2. 创建或补充 `pkg/mathutil/`。
3. 至少实现 7 个函数：
   - `Double`
   - `IsEven`
   - `NormalizeName`
   - `UserLabel`
   - `DoubleAge`
   - `IsAdultAgeEven`
   - `PackageStatus`
4. 至少写 7 组测试。
5. `go list ./...` 可以看到新增 package。
6. `go test ./...` 通过。
7. 能解释：
   - package 和目录的关系
   - module path 和 import path 的关系
   - 导出标识符为什么要大写
   - `init` 什么时候执行

---

## 今天最容易踩的坑

### 坑 1：同一个目录混用多个 package

错误：

```go
// a.go
package packages

// b.go
package mathutil
```

如果它们在同一个目录，通常会编译失败。

---

### 坑 2：导入了但没使用

```go
import "fmt"

func Demo() string {
    return "demo"
}
```

如果没有使用 `fmt`，Go 会编译失败。

---

### 坑 3：小写函数不能跨 package 调用

```go
func double(n int) int {
    return n * 2
}
```

其他 package 不能调用：

```go
mathutil.double(3)
```

要跨 package 调用，需要：

```go
func Double(n int) int {
    return n * 2
}
```

---

### 坑 4：import path 写成文件路径

错误理解：

```go
import "github.com/CanYangTang/go_learning/pkg/mathutil/mathutil.go"
```

正确写法是导入目录路径：

```go
import "github.com/CanYangTang/go_learning/pkg/mathutil"
```

---

### 坑 5：滥用 init

`init` 会自动执行，太多隐藏逻辑会让代码难追踪。

今天只用它做轻量状态初始化，不放复杂逻辑。
