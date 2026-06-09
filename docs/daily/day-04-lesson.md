# Day 04 教案：Go 函数

## 学习目标

学完今天，需要能够做到：

1. 定义和调用普通函数。
2. 理解参数、返回值和多返回值。
3. 使用具名返回值，但知道什么时候不该滥用。
4. 使用函数返回错误信息或状态信息。
5. 使用变参函数处理不固定数量的参数。
6. 理解匿名函数和立即执行函数。
7. 理解闭包如何捕获外部变量。
8. 实现支持加减乘除的计算器函数并通过测试。

---

## 函数是什么

函数是把一段逻辑封装成一个可复用单元。

```go
func Add(a int, b int) int {
    return a + b
}
```

调用：

```go
result := Add(1, 2)
```

函数能带来：

- 复用逻辑
- 拆分复杂问题
- 让代码更容易测试
- 让代码表达意图

---

## 函数基本结构

```go
func 函数名(参数列表) 返回值类型 {
    函数体
}
```

示例：

```go
func Add(a int, b int) int {
    return a + b
}
```

其中：

```text
func        声明函数
Add         函数名
a int       参数 a，类型 int
b int       参数 b，类型 int
int         返回值类型
return      返回结果
```

---

## 参数简写

如果多个连续参数类型相同，可以简写：

```go
func Add(a, b int) int {
    return a + b
}
```

等价于：

```go
func Add(a int, b int) int {
    return a + b
}
```

今天练习里会大量使用这种写法。

---

## 无返回值函数

函数可以没有返回值。

```go
func PrintHello(name string) {
    fmt.Println("hello", name)
}
```

这种函数通常用于：

- 打印
- 写文件
- 修改外部状态
- 启动服务

但业务逻辑函数通常更推荐返回值，这样更容易测试。

---

## 单返回值

```go
func Square(n int) int {
    return n * n
}
```

调用：

```go
value := Square(5)
```

---

## 多返回值

Go 支持函数返回多个值。

```go
func Divide(a, b float64) (float64, bool) {
    if b == 0 {
        return 0, false
    }
    return a / b, true
}
```

调用：

```go
result, ok := Divide(10, 2)
if !ok {
    fmt.Println("divide failed")
}
```

多返回值是 Go 的核心习惯之一，后面错误处理会经常看到：

```go
result, err := doSomething()
if err != nil {
    return err
}
```

今天先用 `(value, ok)` 练习多返回值，Day 5 再正式进入 `error`。

---

## 为什么除法适合多返回值

除法有一个特殊情况：除数不能为 0。

如果只返回一个数字：

```go
func Divide(a, b float64) float64
```

当 `b == 0` 时不好表达失败。

所以可以返回两个值：

```go
func Divide(a, b float64) (float64, bool)
```

含义：

```text
第一个返回值：计算结果
第二个返回值：是否成功
```

例如：

```go
Divide(10, 2) // 5, true
Divide(10, 0) // 0, false
```

---

## 具名返回值

Go 支持给返回值命名。

```go
func Divide(a, b float64) (result float64, ok bool) {
    if b == 0 {
        return 0, false
    }
    return a / b, true
}
```

也可以裸 return：

```go
func Divide(a, b float64) (result float64, ok bool) {
    if b == 0 {
        return
    }
    result = a / b
    ok = true
    return
}
```

但是学习阶段不建议滥用裸 return，因为它容易降低可读性。

建议先写清楚：

```go
return a / b, true
```

---

## 变参函数

变参函数可以接收不固定数量的参数。

```go
func Sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}
```

调用：

```go
Sum(1, 2, 3)    // 6
Sum(1, 2, 3, 4) // 10
Sum()           // 0
```

在函数内部，`nums` 的类型可以理解为：

```go
[]int
```

所以可以用 `for range` 遍历。

---

## for range

`range` 常用于遍历 slice、array、map、string、channel。

今天先看 slice：

```go
nums := []int{1, 2, 3}

for index, value := range nums {
    fmt.Println(index, value)
}
```

如果不需要下标，用 `_` 忽略：

```go
for _, value := range nums {
    fmt.Println(value)
}
```

这会在变参函数里用到。

---

## 匿名函数

匿名函数就是没有名字的函数。

```go
double := func(n int) int {
    return n * 2
}

result := double(5)
```

这里 `double` 是一个变量，保存了一个函数。

函数在 Go 里是一等公民，可以：

- 赋值给变量
- 作为参数传入
- 作为返回值返回

---

## 立即执行匿名函数

匿名函数可以定义后立即执行。

```go
result := func(a, b int) int {
    return a + b
}(1, 2)
```

最后的 `(1, 2)` 表示立即调用这个匿名函数。

这种写法适合把一小段初始化逻辑包起来，但不要滥用。

---

## 闭包

闭包是一个函数捕获了外部变量。

```go
func Counter() func() int {
    count := 0

    return func() int {
        count++
        return count
    }
}
```

调用：

```go
next := Counter()
fmt.Println(next()) // 1
fmt.Println(next()) // 2
fmt.Println(next()) // 3
```

这里内部匿名函数捕获了外部的 `count`，所以每次调用都能记住之前的值。

闭包常用于：

- 生成器
- 状态封装
- 中间件
- 回调

后面学 Gin 中间件时会再次遇到闭包。

---

## 函数命名习惯

Go 函数命名通常使用驼峰：

```go
Add
Divide
FormatScore
NewUserService
```

首字母大小写有特殊含义：

```text
Add      可被其他 package 访问，导出函数
add      只能在当前 package 内访问，非导出函数
```

今天练习函数用大写开头，是为了测试文件和后续示例更直观。

---

## 返回值和测试

函数越纯粹，越容易测试。

容易测试：

```go
func Add(a, b int) int {
    return a + b
}
```

不太方便测试：

```go
func PrintAdd(a, b int) {
    fmt.Println(a + b)
}
```

因为返回值可以直接断言，而打印输出需要捕获 stdout。

所以今天的计算器函数都设计成返回值。

---

## 今日代码练习设计

今天会写一个小 package：

```text
week01-basics/day04-functions/
├── calculator.go
└── calculator_test.go
```

建议实现这些函数：

### 加法

```go
func Add(a, b float64) float64
```

期望：

```go
Add(1, 2) // 3
```

练习点：

- 基础函数定义
- 参数和返回值

### 减法

```go
func Subtract(a, b float64) float64
```

期望：

```go
Subtract(5, 2) // 3
```

### 乘法

```go
func Multiply(a, b float64) float64
```

期望：

```go
Multiply(3, 4) // 12
```

### 除法

```go
func Divide(a, b float64) (float64, bool)
```

期望：

```go
Divide(10, 2) // 5, true
Divide(10, 0) // 0, false
```

练习点：

- 多返回值
- 用 `bool` 表示是否成功

### 变参加法

```go
func Sum(nums ...float64) float64
```

期望：

```go
Sum(1, 2, 3) // 6
Sum()        // 0
```

练习点：

- 变参函数
- `for range`

### 应用运算函数

```go
func Apply(a, b float64, op func(float64, float64) float64) float64
```

期望：

```go
Apply(2, 3, Add) // 5
```

练习点：

- 函数作为参数
- 匿名函数

### 计数器闭包

```go
func NewCounter() func() int
```

期望：

```go
counter := NewCounter()
counter() // 1
counter() // 2
```

练习点：

- 闭包
- 捕获外部变量

---

## 今天最容易踩的坑

### 坑 1：多返回值必须一起接收或明确忽略

```go
result := Divide(10, 2) // 错误
```

正确：

```go
result, ok := Divide(10, 2)
```

如果暂时不需要 ok：

```go
result, _ := Divide(10, 2)
```

但不要滥用 `_`，因为 `ok` 往往表达重要状态。

### 坑 2：除以 0 要显式处理

```go
func Divide(a, b float64) (float64, bool) {
    if b == 0 {
        return 0, false
    }
    return a / b, true
}
```

今天暂时用 `bool`，Day 5 会改成 `error`。

### 坑 3：变参在函数内部像 slice

```go
func Sum(nums ...int) int {
    for _, n := range nums {
        _ = n
    }
    return 0
}
```

`nums` 可以当作 `[]int` 遍历。

### 坑 4：闭包捕获的是变量，不只是值

```go
func Counter() func() int {
    count := 0
    return func() int {
        count++
        return count
    }
}
```

每次调用返回的匿名函数，都会继续使用同一个 `count`。

### 坑 5：函数名首字母大小写影响可见性

```go
func Add(...) // 导出，其他 package 可访问
func add(...) // 非导出，只能当前 package 使用
```

后面拆分 package 时这个规则会很重要。

---

## 今日验收标准

今天完成后应该满足：

1. 创建 `week01-basics/day04-functions/`。
2. 实现基础计算器函数：`Add`、`Subtract`、`Multiply`、`Divide`。
3. 实现至少一个可选函数：`Sum`、`Apply` 或 `NewCounter`。
4. 至少写 6 组测试。
5. `go test ./...` 通过。
6. 能解释：
   - 参数和返回值
   - 多返回值如何接收
   - 变参函数内部为什么像 slice
   - 匿名函数和函数作为参数
   - 闭包如何保存状态
7. 更新 `docs/daily/day-04.md`。
