# Day 07 教案：Week 1 总结与工具包练习

## 学习目标

学完今天，需要能够做到：

1. 复盘 Week 1 学过的 Go 基础知识。
2. 把零散函数整理成可复用 utility package。
3. 理解什么时候适合创建 `pkg/` 下的公共包。
4. 写出清晰的小函数和对应测试。
5. 使用 `make fmt`、`make test`、`make vet` 做完整验证。
6. 编写 Week 1 周总结，明确掌握点、薄弱点和下一周目标。

---

## Day 7 的定位

Day 1 到 Day 6 已经覆盖了 Go 入门阶段最核心的基础：

- 项目初始化和基础工程结构。
- 变量、常量、基础类型、类型转换。
- 条件、循环、`switch`。
- 函数、多返回值、变参、匿名函数、闭包。
- `error`、错误包装、`defer`、`panic/recover`。
- `package`、`import`、module path、`init`。

Day 7 不急着学习大量新语法，而是做两件事：

1. 把已经学过的内容串起来。
2. 用一个小工具包练习“可复用代码 + 测试 + 文档化复盘”。

这比继续堆新知识更重要，因为基础阶段最容易出现的问题不是“没见过语法”，而是“见过但不能组合使用”。

---

## 为什么要做周总结

周总结不是写流水账，而是把学习内容变成可复用的知识结构。

好的周总结应该回答：

1. 这一周我学了哪些主题？
2. 哪些概念已经能独立解释？
3. 哪些概念还容易混淆？
4. 哪些代码是我能独立写出来的？
5. 下一周需要重点练什么？

对于工程学习来说，周总结还能帮助你回头看：

- 项目结构是否越来越清晰。
- 测试是否覆盖了核心行为。
- 错误是否被明确表达。
- package 拆分是否合理。

---

## utility package 是什么

utility package 通常是一些通用、独立、可复用的小工具函数集合。

例如：

```text
pkg/stringutil/
├── stringutil.go
└── stringutil_test.go
```

它可能包含：

```go
func NormalizeSpace(s string) string
func IsBlank(s string) bool
func JoinNonEmpty(parts ...string) string
```

这些函数不依赖业务状态，不连接数据库，不调用网络，只做明确的小逻辑。

---

## 什么时候适合放到 pkg

适合放到 `pkg/` 的代码通常有这些特点：

- 可被多个 package 使用。
- 不依赖具体业务流程。
- 输入输出明确。
- 容易测试。
- 名字能清楚表达用途。

例如当前已有：

```text
pkg/mathutil/
```

这是合理的，因为数学工具函数可以被不同练习或业务代码复用。

---

## 什么时候不应该急着抽工具包

不要为了“看起来工程化”而过早抽象。

不适合放进工具包的情况：

- 只有一个地方使用。
- 逻辑强依赖某个业务流程。
- 名字很模糊，例如 `helper`、`common`。
- 只是为了少写两行代码。
- 还没弄清楚复用边界。

学习阶段可以小规模练习工具包，但要记住：工具包应该是“自然长出来”的，而不是强行拆出来的。

---

## 补充专题：slice 基础

你前面已经接触过 `[]int`、`[]string` 这样的写法，但还没有系统学习 slice。今天的 `JoinNonEmpty` 会用到 slice，所以先补这部分。

---

## slice 是什么

slice 是 Go 中最常用的序列类型，可以理解为“动态长度的列表”。

例如：

```go
nums := []int{1, 2, 3}
names := []string{"Alice", "Bob"}
```

这里：

```go
[]int
[]string
```

表示 slice 类型。

slice 和数组不同：

```go
arr := [3]int{1, 2, 3} // 数组，长度是类型的一部分
s := []int{1, 2, 3}   // slice，长度可以变化
```

日常 Go 代码中更常用 slice。

---

## 声明空 slice

常见写法：

```go
var nums []int
```

这会得到一个 nil slice，长度是 0。

也可以写：

```go
nums := []int{}
```

这会得到一个空 slice，长度也是 0。

对学习阶段来说，重点记住：空 slice 可以安全 `append`。

```go
var nums []int
nums = append(nums, 1)
nums = append(nums, 2)
```

---

## append

`append` 用来向 slice 追加元素。

```go
nums := []int{1, 2}
nums = append(nums, 3)
```

注意：`append` 会返回新的 slice，所以必须接住返回值。

正确：

```go
nums = append(nums, 3)
```

错误：

```go
append(nums, 3)
```

这也是 Day 5 `DeferOrder` 里容易踩坑的点。

---

## len

`len` 可以获取 slice 长度。

```go
nums := []int{1, 2, 3}
fmt.Println(len(nums)) // 3
```

空 slice 的长度是 0。

```go
var nums []int
fmt.Println(len(nums)) // 0
```

---

## for range 遍历 slice

常见写法：

```go
nums := []int{1, 2, 3}

for index, value := range nums {
    fmt.Println(index, value)
}
```

如果不需要下标，可以用 `_` 忽略：

```go
for _, value := range nums {
    fmt.Println(value)
}
```

今天的 `JoinNonEmpty` 会遍历 `parts ...string`，在函数内部它可以当作 `[]string` 使用。

---

## 变参和 slice 的关系

函数：

```go
func JoinNonEmpty(sep string, parts ...string) string
```

这里的 `parts ...string` 是变参。

在函数内部，`parts` 的类型可以当成：

```go
[]string
```

所以可以：

```go
for _, part := range parts {
    // 使用 part
}
```

如果你已经有一个 slice，调用变参函数时可以用 `...` 展开：

```go
items := []string{"a", "b"}
JoinNonEmpty(",", items...)
```

---

## slice 在今天练习中的用途

`JoinNonEmpty` 的核心思路是：

1. 创建一个临时 slice 保存清理后的非空字符串。
2. 遍历所有传入参数。
3. 对每个字符串做 `NormalizeSpace`。
4. 如果不是空字符串，就 append 到临时 slice。
5. 最后用 `strings.Join` 拼接。

伪代码：

```go
var cleaned []string

for _, part := range parts {
    normalized := NormalizeSpace(part)
    if normalized == "" {
        continue
    }
    cleaned = append(cleaned, normalized)
}

return strings.Join(cleaned, sep)
```

这段代码会同时练习 slice、`for range`、`if`、`continue`、变参和标准库。

---

## 今天的工具包练习

今天建议新增：

```text
pkg/stringutil/
├── stringutil.go
└── stringutil_test.go
```

实现这些函数：

```go
func IsBlank(s string) bool
func NormalizeSpace(s string) string
func JoinNonEmpty(sep string, parts ...string) string
```

它们会综合练习：

- `package`
- `import`
- 字符串处理
- `for range`
- `if`
- 变参函数
- slice
- 测试

---

## IsBlank

目标：判断一个字符串去掉空白后是否为空。

期望：

```go
IsBlank("")        // true
IsBlank("   ")     // true
IsBlank(" Alice ") // false
```

可以用：

```go
strings.TrimSpace(s)
```

如果裁剪后等于空字符串，说明它是 blank。

---

## NormalizeSpace

目标：去掉字符串两边空白，并把中间连续空白压缩成一个空格。

期望：

```go
NormalizeSpace("  Alice   Bob  ") // "Alice Bob"
```

可以使用：

```go
strings.Fields(s)
strings.Join(fields, " ")
```

`strings.Fields` 会按连续空白切分字符串，并自动忽略前后空白。

示例：

```go
fields := strings.Fields("  Alice   Bob  ")
// []string{"Alice", "Bob"}
```

再 Join：

```go
strings.Join(fields, " ")
// "Alice Bob"
```

---

## JoinNonEmpty

目标：只拼接非空内容。

函数签名：

```go
func JoinNonEmpty(sep string, parts ...string) string
```

期望：

```go
JoinNonEmpty(",", "a", "", "b") // "a,b"
JoinNonEmpty("-", "go", "", "lang") // "go-lang"
JoinNonEmpty(",", "", " ") // ""
```

建议规则：

1. 遍历 `parts`。
2. 对每个 part 先 `NormalizeSpace`。
3. 如果结果不是空字符串，就加入临时 slice。
4. 最后用 `strings.Join(cleaned, sep)` 拼接。

这个函数能综合练习：

- 变参函数。
- slice 追加。
- 调用同 package 内其他函数。
- 标准库 `strings`。

---

## 测试应该测什么

工具函数测试要覆盖：

- 正常输入。
- 空字符串。
- 只有空白的字符串。
- 多个连续空白。
- 变参为空。
- 部分参数为空。

例如：

```go
func TestIsBlank(t *testing.T) {
    tests := []struct {
        name string
        input string
        want bool
    }{
        {name: "empty", input: "", want: true},
        {name: "spaces", input: "   ", want: true},
        {name: "text", input: " Alice ", want: false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := IsBlank(tt.input)
            if got != tt.want {
                t.Fatalf("IsBlank() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

---

## Week 1 周总结应该怎么写

今天还要更新：

```text
docs/weekly/week-01.md
```

建议结构：

```markdown
# Week 01 总结：Go Foundations

## 本周目标

## 已完成内容

## 代码产出

## 掌握较好的知识点

## 仍需巩固的知识点

## 典型错误与修正

## 验证命令

## 下周计划
```

这不是写给别人看的汇报，而是写给未来的自己看的复盘。

---

## 本周知识地图

可以按主题整理：

### 工程基础

- `go.mod`
- module path
- Makefile
- `go test ./...`
- `go fmt ./...`
- `go vet ./...`

### 语法基础

- 变量声明
- 常量
- 零值
- 基础类型
- 显式类型转换

### 控制流

- `if/else`
- `for`
- `switch`
- `break`
- `continue`

### 函数

- 参数
- 返回值
- 多返回值
- 变参
- 函数作为参数
- 匿名函数
- 闭包

### 错误处理

- `error`
- `errors.New`
- sentinel error
- `fmt.Errorf`
- `%w`
- `errors.Is`
- `defer`
- `panic/recover`

### 包管理

- `package`
- `import`
- 导出标识符
- module path
- import path
- `init`

---

## 常见复盘误区

### 误区 1：只写“学了什么”

更好的方式是写：

- 学了什么。
- 用代码实现了什么。
- 哪些测试验证了它。
- 哪些地方还不熟。

---

### 误区 2：只记录成功，不记录错误

错误很重要，因为它们暴露了真正需要巩固的地方。

例如本周已经出现过这些典型混淆：

- `fmt.Sprintf` 和 `fmt.Printf`。
- `string(123)` 和 `strconv.Itoa(123)`。
- 函数作为参数和闭包。
- 匿名函数作为参数和立即执行。
- `err == nil` 与 `err != nil`。
- `defer` 修改局部变量和具名返回值的区别。

这些都应该写进周总结。

---

### 误区 3：写得太泛

不建议只写：

```markdown
本周学习了 Go 基础。
```

建议写成：

```markdown
本周能独立写出带测试的基础函数，能使用 `go test ./...` 验证整个 module，能用 `errors.Is` 判断被 `%w` 包装过的 sentinel error。
```

越具体，越有复盘价值。

---

## 今日验收标准

完成后应该满足：

1. 创建 `pkg/stringutil/`。
2. 实现 `IsBlank`、`NormalizeSpace`、`JoinNonEmpty`。
3. 为 `pkg/stringutil` 写测试。
4. 更新 `docs/weekly/week-01.md`。
5. `make fmt`、`make test`、`make vet` 通过。
6. 能解释本周核心知识地图。

---

## 明天进入什么内容

Day 8 将进入 Week 2：并发基础。

重点会开始学习：

- goroutine
- 并发执行
- `time.Sleep`
- `sync.WaitGroup`
- 并发中的输出顺序

所以 Day 7 的目的，是在进入并发前把 Week 1 的基础打牢。
