# Day 02 教案：Go 变量、常量、基础类型和类型转换

## 学习目标

学完今天，需要能够做到：

1. 知道 Go 中变量的几种声明方式。
2. 理解 `var`、`:=`、`const` 的区别。
3. 掌握基础类型：`int`、`float64`、`string`、`bool`。
4. 理解 Go 的零值机制。
5. 理解 Go 不支持隐式类型转换。
6. 正确进行数字和字符串之间的转换。
7. 能写简单函数并用测试验证。

---

## 变量是什么

变量可以理解为：给一块数据起一个名字，之后通过这个名字使用它。

```go
var age int = 18
```

含义：

```text
定义一个变量 age
它的类型是 int
它的值是 18
```

之后可以使用：

```go
fmt.Println(age)
```

---

## Go 的变量声明方式

### 显式声明类型

```go
var age int = 18
```

特点：

- 类型写得很清楚。
- 适合学习阶段，帮助建立类型意识。
- 也适合需要明确类型的场景，比如 `int64`、`float64`。

示例：

```go
var name string = "Alice"
var age int = 18
var height float64 = 1.75
var active bool = true
```

### 类型推断

```go
var age = 18
```

Go 会根据右边的值推断类型。

```go
var name = "Alice"   // string
var age = 18         // int
var height = 1.75    // float64
var active = true    // bool
```

注意：

```go
var height = 1.75 // 默认推断为 float64
var age = 18      // 默认推断为 int
```

### 短变量声明

```go
age := 18
```

这是 Go 中非常常见的写法，等价于：

```go
var age = 18
```

但有一个重要限制：`:=` 只能在函数内部使用。

可以：

```go
func main() {
    age := 18
    fmt.Println(age)
}
```

不可以：

```go
age := 18 // package 级别不能这样写

func main() {}
```

package 级别应该写：

```go
var age = 18
```

### 先声明，后赋值

```go
var age int
age = 18
```

如果只声明不赋值，Go 会给变量一个默认值，叫做零值。

```go
var age int
fmt.Println(age) // 0
```

---

## Go 的零值

Go 中变量只声明不赋值时，会自动获得零值。

| 类型 | 零值 |
|---|---|
| `int` | `0` |
| `float64` | `0` |
| `string` | `""` 空字符串 |
| `bool` | `false` |
| pointer | `nil` |
| slice | `nil` |
| map | `nil` |
| channel | `nil` |
| interface | `nil` |

示例：

```go
var age int
var price float64
var name string
var ok bool

fmt.Println(age)   // 0
fmt.Println(price) // 0
fmt.Println(name)  // ""
fmt.Println(ok)    // false
```

零值是 Go 很重要的设计。它让很多类型在声明后就有一个可预测状态。

---

## 常量 const

常量用于表示不会变化的值。

```go
const AppName = "go-learning"
const Pi = 3.14159
```

常量不能被重新赋值：

```go
const AppName = "go-learning"
AppName = "new-name" // 编译错误
```

### 常量和变量的区别

| 对比 | 变量 var | 常量 const |
|---|---|---|
| 是否可变 | 可以重新赋值 | 不能重新赋值 |
| 使用场景 | 会变化的数据 | 固定值、配置名、数学常量 |
| 示例 | 年龄、用户名、请求参数 | 版本号、状态码、固定文案 |

---

## 基础类型

今天先掌握 4 个基础类型。

### int

整数类型。

```go
var age int = 18
var count int = 100
```

常见用途：

- 年龄
- 数量
- 下标
- 次数
- 页码

注意：

```go
var a int = 10
var b int64 = 20
```

`int` 和 `int64` 是不同类型，不能直接相加：

```go
a + b // 编译错误
```

必须转换：

```go
int64(a) + b
```

### float64

浮点数类型。

```go
var price float64 = 19.99
var rate float64 = 0.85
```

常见用途：

- 比率
- 平均值
- 面积
- 分数

学习阶段可以用 `float64` 表示金额，但真实金融业务通常不用浮点数直接存金额。

示例：

```go
var total int = 95
var count int = 10

avg := float64(total) / float64(count)
fmt.Println(avg) // 9.5
```

如果写：

```go
avg := total / count
```

结果是整数除法：

```text
9
```

不是：

```text
9.5
```

### string

字符串类型。

```go
var name string = "Alice"
```

字符串可以拼接：

```go
message := "hello, " + name
```

也可以格式化：

```go
fmt.Sprintf("name=%s age=%d", name, age)
```

常见占位符：

| 占位符 | 含义 |
|---|---|
| `%s` | string |
| `%d` | int |
| `%f` | float |
| `%.2f` | 保留 2 位小数 |
| `%t` | bool |
| `%v` | 通用值 |

示例：

```go
name := "Alice"
age := 18
score := 95.5

text := fmt.Sprintf("%s is %d years old, score=%.1f", name, age, score)
```

### bool

布尔类型，只有两个值：

```go
true
false
```

示例：

```go
var active bool = true
var deleted bool = false
```

常用于条件判断：

```go
age := 18
isAdult := age >= 18

if isAdult {
    fmt.Println("adult")
}
```

---

## 类型转换

Go 是强类型语言，不会自动帮你做隐式类型转换。

### 数字之间转换

```go
var age int = 18
var score float64 = 95.5

result := float64(age) + score
```

如果不转换：

```go
result := age + score
```

会编译错误。

### int 转 string

错误理解：

```go
string(123)
```

这个不是得到 `"123"`，而是把 `123` 当作 Unicode 码点，得到对应字符。

正确方式：

```go
strconv.Itoa(123)
```

示例：

```go
import "strconv"

age := 18
text := strconv.Itoa(age)
```

### string 转 int

```go
import "strconv"

ageText := "18"
age, err := strconv.Atoi(ageText)
if err != nil {
    // 处理转换失败
}
```

为什么有 `err`？因为字符串可能不是合法数字：

```go
strconv.Atoi("abc")
```

这个无法转换成整数，所以会返回错误。

### float64 转 string

```go
strconv.FormatFloat(3.14159, 'f', 2, 64)
```

含义：

```text
3.14159       要转换的浮点数
'f'           普通小数格式
2             保留 2 位小数
64            float64
```

结果：

```go
"3.14"
```

也可以用：

```go
fmt.Sprintf("%.2f", 3.14159)
```

### string 转 float64

```go
price, err := strconv.ParseFloat("19.99", 64)
if err != nil {
    // 处理转换失败
}
```

---

## `fmt.Sprintf` 和 `strconv` 的区别

两者都能把值变成字符串，但用途不同。

### fmt.Sprintf

适合组装文本：

```go
fmt.Sprintf("name=%s age=%d", name, age)
```

优点：

- 可读性好
- 适合多字段拼接
- 适合输出给人看

### strconv

适合明确的类型转换：

```go
strconv.Itoa(age)
strconv.Atoi("18")
strconv.FormatFloat(price, 'f', 2, 64)
strconv.ParseFloat("19.99", 64)
```

优点：

- 转换语义明确
- 性能通常更好
- 适合程序内部转换

简单记：

```text
拼一句话：fmt.Sprintf
做类型转换：strconv
```

---

## Go 中未使用变量会编译失败

Go 不允许声明变量但不使用。

错误示例：

```go
func main() {
    age := 18
}
```

如果 `age` 没有被使用，编译会失败。

临时不想用可以写：

```go
_ = age
```

但不要滥用。学习时可以临时用，项目代码里应该尽量删除无用变量。

---

## 命名习惯

Go 里变量命名通常使用小驼峰：

```go
userName
totalCount
isActive
```

不推荐使用下划线风格：

```go
user_name // 不推荐
```

局部变量可以短一些：

```go
i := 0
n := len(items)
err := doSomething()
```

---

## 今日代码练习设计

今天会写一个小 package：

```text
week01-basics/day02-syntax/
├── syntax.go
└── syntax_test.go
```

### 用户信息格式化

```go
func UserProfile(name string, age int) string
```

期望：

```go
UserProfile("Alice", 18)
```

返回：

```text
Alice is 18 years old
```

练习点：

- `string`
- `int`
- `fmt.Sprintf`

### 矩形面积

```go
func RectangleArea(width, height float64) float64
```

期望：

```go
RectangleArea(3, 4)
```

返回：

```go
12
```

练习点：

- `float64`
- 多个参数同类型可以简写

### 平均值

```go
func Average(total int, count int) float64
```

期望：

```go
Average(95, 10)
```

返回：

```go
9.5
```

练习点：

- `int`
- `float64`
- 显式类型转换
- 避免整数除法

### 判断成年人

```go
func IsAdult(age int) bool
```

期望：

```go
IsAdult(18) // true
IsAdult(17) // false
```

练习点：

- `bool`
- 比较表达式

### 分数格式化

```go
func FormatScore(score int) string
```

期望：

```go
FormatScore(95)
```

返回：

```text
score=95
```

练习点：

- `int` 转 `string`
- `strconv.Itoa` 或 `fmt.Sprintf`

---

## 今天最容易踩的坑

### 坑 1：`int` 和 `float64` 不能直接运算

错误：

```go
var a int = 10
var b float64 = 2.5
fmt.Println(a + b)
```

正确：

```go
fmt.Println(float64(a) + b)
```

### 坑 2：整数除法会丢小数

```go
total := 95
count := 10
avg := total / count
```

结果是：

```go
9
```

不是：

```go
9.5
```

正确：

```go
avg := float64(total) / float64(count)
```

### 坑 3：`string(123)` 不是 `"123"`

错误理解：

```go
text := string(123)
```

正确：

```go
text := strconv.Itoa(123)
```

### 坑 4：声明了变量但没用会编译失败

错误：

```go
func main() {
    age := 18
}
```

正确：

```go
func main() {
    age := 18
    fmt.Println(age)
}
```

---

## 今日验收标准

今天完成后应该满足：

1. 创建 `week01-basics/day02-syntax/`。
2. 至少实现 5 个基础函数。
3. 至少写 5 个测试。
4. `go test ./...` 通过。
5. 能解释：
   - `var` 和 `:=` 的区别
   - 零值是什么
   - 为什么 Go 需要显式类型转换
   - `string(123)` 为什么不是 `"123"`
6. 更新 `docs/daily/day-02.md`。
