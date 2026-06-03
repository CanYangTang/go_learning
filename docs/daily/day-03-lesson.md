# Day 03 教案：Go 流程控制

## 学习目标

学完今天，需要能够做到：

1. 使用 `if/else` 编写条件分支。
2. 理解 Go 中 `if` 可以带初始化语句。
3. 使用 `for` 完成计数循环、条件循环和无限循环。
4. 使用 `break` 和 `continue` 控制循环流程。
5. 使用 `switch` 编写多分支逻辑。
6. 实现九九乘法表。
7. 编写基础函数并用测试验证流程控制逻辑。

---

## 流程控制是什么

流程控制决定程序执行路径。

常见问题包括：

- 满足某个条件时执行什么？
- 不满足条件时执行什么？
- 一段逻辑要重复执行多少次？
- 多个可能值应该走哪个分支？

Go 中今天重点学习：

```go
if / else
for
switch
break
continue
```

---

## if / else

`if` 用于条件判断。

```go
age := 18

if age >= 18 {
    fmt.Println("adult")
} else {
    fmt.Println("minor")
}
```

Go 的 `if` 条件不需要小括号，但代码块必须使用花括号。

错误写法：

```go
if (age >= 18)
    fmt.Println("adult")
```

正确写法：

```go
if age >= 18 {
    fmt.Println("adult")
}
```

---

## 多分支 if / else if / else

```go
score := 85

if score >= 90 {
    fmt.Println("A")
} else if score >= 80 {
    fmt.Println("B")
} else if score >= 60 {
    fmt.Println("C")
} else {
    fmt.Println("D")
}
```

分支会从上到下判断，一旦命中某个分支，后面的分支就不会继续执行。

因此条件顺序很重要。

例如成绩等级应该从高到低判断：

```go
if score >= 90 {
    return "A"
} else if score >= 80 {
    return "B"
}
```

如果先写：

```go
if score >= 60 {
    return "C"
} else if score >= 90 {
    return "A"
}
```

那么 95 分也会先命中 `score >= 60`，不会得到 A。

---

## if 初始化语句

Go 的 `if` 可以在条件前写一个初始化语句。

```go
if score := 85; score >= 60 {
    fmt.Println("pass")
}
```

这里的 `score` 只在 `if/else` 结构内有效。

常见于错误处理：

```go
if err := doSomething(); err != nil {
    return err
}
```

这是一种 Go 中非常常见的写法。

---

## for 循环

Go 只有一种循环关键字：`for`。

### 计数循环

```go
for i := 1; i <= 5; i++ {
    fmt.Println(i)
}
```

结构：

```text
for 初始化; 条件; 更新 {
    循环体
}
```

执行顺序：

1. 执行初始化：`i := 1`
2. 判断条件：`i <= 5`
3. 执行循环体
4. 执行更新：`i++`
5. 回到第 2 步

---

## 条件循环

Go 的 `for` 也可以像其他语言里的 `while`。

```go
count := 3

for count > 0 {
    fmt.Println(count)
    count--
}
```

Go 没有 `while` 关键字，条件循环也用 `for`。

---

## 无限循环

```go
for {
    fmt.Println("running")
}
```

一般需要配合 `break` 或外部退出条件，否则会一直运行。

---

## break

`break` 用于退出当前循环。

```go
for i := 1; i <= 10; i++ {
    if i == 5 {
        break
    }
    fmt.Println(i)
}
```

输出：

```text
1
2
3
4
```

当 `i == 5` 时，循环直接结束。

---

## continue

`continue` 用于跳过本轮循环，进入下一轮。

```go
for i := 1; i <= 5; i++ {
    if i == 3 {
        continue
    }
    fmt.Println(i)
}
```

输出：

```text
1
2
4
5
```

当 `i == 3` 时，跳过本轮后续代码。

---

## 嵌套循环

九九乘法表需要用嵌套循环。

```go
for i := 1; i <= 9; i++ {
    for j := 1; j <= i; j++ {
        fmt.Printf("%d*%d=%d ", j, i, i*j)
    }
    fmt.Println()
}
```

外层循环控制行，内层循环控制每一行的列。

例如：

```text
1*1=1
1*2=2 2*2=4
1*3=3 2*3=6 3*3=9
```

---

## switch

`switch` 用于多分支选择。

```go
level := "A"

switch level {
case "A":
    fmt.Println("excellent")
case "B":
    fmt.Println("good")
case "C":
    fmt.Println("pass")
default:
    fmt.Println("unknown")
}
```

Go 的 `switch` 默认不会继续执行下一个 case，不需要写 `break`。

这和很多语言不一样。

---

## switch 多个匹配值

```go
weekday := "Sat"

switch weekday {
case "Sat", "Sun":
    fmt.Println("weekend")
default:
    fmt.Println("weekday")
}
```

多个 case 值可以写在同一个 case 中。

---

## switch 不带表达式

`switch` 可以不带变量，直接写条件。

```go
score := 85

switch {
case score >= 90:
    fmt.Println("A")
case score >= 80:
    fmt.Println("B")
case score >= 60:
    fmt.Println("C")
default:
    fmt.Println("D")
}
```

这种写法适合替代复杂的 `if/else if`。

---

## switch 和 if 怎么选

简单记：

```text
范围判断：if / else if
固定值匹配：switch
多个条件分支：switch 不带表达式也可以
```

例子：

成绩等级这种范围判断：

```go
if score >= 90 { ... }
```

菜单选项这种固定值匹配：

```go
switch option {
case "1":
case "2":
}
```

---

## 常见坑

### 坑 1：if 条件不能是非 bool

Go 中 `if` 条件必须是 `bool`。

错误：

```go
count := 1
if count {
    fmt.Println("ok")
}
```

正确：

```go
if count > 0 {
    fmt.Println("ok")
}
```

### 坑 2：for 循环边界容易错

```go
for i := 1; i < 9; i++ {
}
```

这个只会到 8。

九九乘法表需要到 9：

```go
for i := 1; i <= 9; i++ {
}
```

### 坑 3：switch 默认不会贯穿

Go 中：

```go
switch n {
case 1:
    fmt.Println("one")
case 2:
    fmt.Println("two")
}
```

命中 `case 1` 后不会继续执行 `case 2`。

如果确实需要贯穿，要使用 `fallthrough`，但日常很少用。

### 坑 4：continue 只跳过当前这一轮

`continue` 不是结束循环，而是进入下一轮。

结束循环应该用 `break`。

### 坑 5：嵌套循环中 break 默认只跳出当前层

```go
for i := 0; i < 3; i++ {
    for j := 0; j < 3; j++ {
        break
    }
}
```

这里的 `break` 只跳出内层循环，不会跳出外层循环。

---

## 今日代码练习设计

今天会写一个小 package：

```text
week01-basics/day03-control-flow/
├── control_flow.go
└── control_flow_test.go
```

建议实现这些函数：

### 判断成绩等级

```go
func Grade(score int) string
```

规则：

```text
90-100: A
80-89: B
60-79: C
0-59: D
其他值: Invalid
```

练习点：

- `if/else if/else`
- 条件顺序
- 边界判断

### 判断工作日/周末

```go
func DayType(day string) string
```

规则：

```text
Mon/Tue/Wed/Thu/Fri: weekday
Sat/Sun: weekend
其他值: unknown
```

练习点：

- `switch`
- 一个 case 匹配多个值
- `default`

### 生成 1 到 n 的偶数

```go
func EvenNumbers(n int) []int
```

示例：

```go
EvenNumbers(6) // []int{2, 4, 6}
```

练习点：

- `for`
- `continue`
- slice 追加

### 求 1 到 n 的和

```go
func SumTo(n int) int
```

示例：

```go
SumTo(5) // 15
```

练习点：

- `for`
- 累加器变量

### 生成九九乘法表

```go
func MultiplicationTable() []string
```

返回 9 行字符串：

```text
1*1=1
1*2=2 2*2=4
...
1*9=9 2*9=18 ... 9*9=81
```

练习点：

- 嵌套循环
- `fmt.Sprintf`
- 字符串拼接或 `strings.Join`

---

## 今日验收标准

今天完成后应该满足：

1. 创建 `week01-basics/day03-control-flow/`。
2. 至少实现 5 个流程控制函数。
3. 至少写 5 组测试。
4. `go test ./...` 通过。
5. 能解释：
   - `if/else` 适合什么场景
   - Go 中 `for` 的几种写法
   - `break` 和 `continue` 的区别
   - Go 的 `switch` 为什么通常不需要 `break`
   - 嵌套循环如何生成九九乘法表
6. 更新 `docs/daily/day-03.md`。
