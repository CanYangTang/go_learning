# Day 13 教案：文件操作

## 学习目标

学完今天，需要能够做到：

1. 理解 `os` 包的文件打开、创建、关闭操作。
2. 理解 `os.Open`、`os.Create`、`os.ReadFile`、`os.WriteFile` 的用法。
3. 理解 `bufio` 包的缓冲读写优势。
4. 理解 `io` 包的通用读写接口。
5. 实现简单的文件读取、写入、按行读取。
6. 编写文件操作的测试。
7. 理解文件路径、相对路径、绝对路径。

---

## Day 13 的位置

Day 12 学了 JSON 请求和响应处理。

Day 13 会把重点放在文件读写，为后续 TODO API 的数据持久化做准备。

Day 14 会做 HTTP 并发综合练习。

---

## 文件操作是什么

文件操作包括：

- 打开现有文件。
- 创建新文件。
- 读取文件内容。
- 写入内容到文件。
- 关闭文件。

Go 标准库提供了多种方式：

- `os` 包：底层文件操作。
- `bufio` 包：缓冲读写，提高性能。
- `io` 包：通用读写接口和工具函数。
- `io/ioutil` 包：简单的一步读写（Go 1.16 后合并到 `os`）。

---

## `os.Open`

`os.Open` 打开一个现有文件，只读模式。

```go
file, err := os.Open("data.txt")
if err != nil {
    // 处理错误，比如文件不存在
}
defer file.Close()
```

返回的 `file` 是 `*os.File` 类型，实现了 `io.Reader` 接口。

注意：打开后要 `defer file.Close()`。

---

## `os.Create`

`os.Create` 创建一个新文件，如果文件已存在会清空内容。

```go
file, err := os.Create("output.txt")
if err != nil {
    // 处理错误，比如目录不存在
}
defer file.Close()
```

返回的 `file` 是 `*os.File` 类型，实现了 `io.Writer` 接口。

---

## `os.ReadFile`

Go 1.16 提供的简单函数，一次性读取整个文件内容。

```go
data, err := os.ReadFile("data.txt")
if err != nil {
    // 处理错误
}
content := string(data)
```

返回的是 `[]byte`，可以直接转成 `string`。

不需要手动打开和关闭文件。

---

## `os.WriteFile`

Go 1.16 提供的简单函数，一次性写入整个文件内容。

```go
content := "hello world"
err := os.WriteFile("output.txt", []byte(content), 0644)
if err != nil {
    // 处理错误
}
```

第三个参数 `0644` 是文件权限：

- `0644`： owner 可读写，group 和 others 可读。
- `0755`：owner 可读写执行，group 和 others 可读执行。

不需要手动创建和关闭文件。

---

## `bufio` 包

`bufio` 提供缓冲读写，减少系统调用次数，提高性能。

### `bufio.NewReader`

```go
file, err := os.Open("data.txt")
if err != nil {
    // 处理错误
}
defer file.Close()

reader := bufio.NewReader(file)
line, err := reader.ReadString('\n')
if err != nil {
    // 处理错误
}
```

`ReadString('\n')` 会读取到换行符为止。

### `bufio.NewScanner`

更常用的按行读取方式：

```go
file, err := os.Open("data.txt")
if err != nil {
    // 处理错误
}
defer file.Close()

scanner := bufio.NewScanner(file)
for scanner.Scan() {
    line := scanner.Text()
    // 处理每一行
}
if err := scanner.Err(); err != nil {
    // 处理扫描错误
}
```

`scanner.Scan()` 会逐行扫描，`scanner.Text()` 返回当前行内容。

### `bufio.NewWriter`

```go
file, err := os.Create("output.txt")
if err != nil {
    // 处理错误
}
defer file.Close()

writer := bufio.NewWriter(file)
writer.WriteString("hello\n")
writer.WriteString("world\n")
writer.Flush() // 必须 Flush，否则缓冲区内容不会写入文件
```

`Flush()` 把缓冲区内容写入文件。

---

## `io` 包

`io` 包定义了 `io.Reader` 和 `io.Writer` 接口。

### `io.Reader`

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

很多类型实现了 `io.Reader`：

- `*os.File`
- `bufio.Reader`
- `bytes.Buffer`
- `strings.Reader`
- `http.Request.Body`

### `io.Writer`

```go
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

很多类型实现了 `io.Writer`：

- `*os.File`
- `bufio.Writer`
- `bytes.Buffer`
- `http.ResponseWriter`

### `io.Copy`

`io.Copy` 从 Reader 复制到 Writer：

```go
src, _ := os.Open("source.txt")
dst, _ := os.Create("dest.txt")
io.Copy(dst, src)
```

---

## 文件路径

Go 可以用相对路径或绝对路径。

相对路径从当前工作目录开始：

```go
os.Open("data.txt")        // 当前目录
os.Open("subdir/data.txt") // 子目录
```

绝对路径从根目录开始：

```go
os.Open("/tmp/data.txt")
os.Open("C:\\temp\\data.txt") // Windows
```

`os.Getwd()` 返回当前工作目录。

---

## `os.Stat`

检查文件是否存在：

```go
info, err := os.Stat("data.txt")
if os.IsNotExist(err) {
    // 文件不存在
}
if err != nil {
    // 其他错误
}
// info.Name(), info.Size(), info.Mode()
```

---

## 今日代码练习设计

今天建议创建：

```text
week02-core/day13-files/
├── reader.go
├── writer.go
├── reader_test.go
└── writer_test.go
```

建议实现：

```go
// reader.go
func ReadAll(path string) (string, error)
func ReadLines(path string) ([]string, error)

// writer.go
func WriteAll(path string, content string) error
func WriteLines(path string, lines []string) error
```

行为：

- `ReadAll`：用 `os.ReadFile` 读取整个文件内容，返回 `string`。
- `ReadLines`：用 `bufio.Scanner` 按行读取，返回每行的 `[]string`。
- `WriteAll`：用 `os.WriteFile` 写入整个内容。
- `WriteLines`：用 `bufio.Writer` 写入多行，每行加换行符。

---

## 测试文件操作

文件测试通常需要：

1. 在测试中创建临时文件。
2. 写入测试数据。
3. 调用被测函数读取或写入。
4. 验证结果。
5. 清理临时文件。

### 使用 `os.CreateTemp`

```go
file, err := os.CreateTemp("", "test*.txt")
if err != nil {
    // 处理错误
}
defer os.Remove(file.Name()) // 测试结束后删除
defer file.Close()

file.WriteString("hello\nworld\n")
```

`os.CreateTemp("", "test*.txt")` 在系统临时目录创建唯一文件名。

### 使用 `os.TempDir`

```go
dir := os.TempDir()
path := filepath.Join(dir, "test.txt")
```

---

## 今日验收标准

完成后应该满足：

1. 创建 `week02-core/day13-files/`。
2. 实现 `ReadAll`、`ReadLines`、`WriteAll`、`WriteLines`。
3. 至少写 4 组测试。
4. `go test ./week02-core/day13-files` 通过。
5. `make test` 通过。
6. 能解释 `os.Open`、`os.Create`、`os.ReadFile`、`os.WriteFile`、`bufio.Scanner`、`bufio.Writer`、`io.Reader`、`io.Writer`。

---

## 可选挑战题：读取 JSON 文件

Issue 的可选挑战是读取 JSON 文件并解析成 struct。

例如：

```go
func ReadJSONFile(path string, v any) error
```

实现思路：

1. 用 `os.ReadFile` 读取整个文件。
2. 用 `json.Unmarshal` 解析到 `v`。
3. 返回错误。

---

## 今天最容易踩的坑

### 坑 1：忘记关闭文件

打开文件后必须关闭，否则会占用系统资源。

```go
file, err := os.Open("data.txt")
if err != nil {
    return err
}
defer file.Close() // 必须关闭
```

---

### 坑 2：忘记 `Flush`

使用 `bufio.Writer` 时，必须调用 `Flush()`。

否则缓冲区内容可能不会写入文件。

```go
writer := bufio.NewWriter(file)
writer.WriteString("hello")
writer.Flush() // 必须 Flush
```

---

### 坑 3：文件路径错误

测试时使用相对路径可能因为工作目录不同而失败。

建议用 `os.CreateTemp` 或 `os.TempDir` 创建临时文件。

---

### 坑 4：忽略 `os.IsNotExist`

检查文件是否存在时，不要只判断 `err != nil`。

应该用 `os.IsNotExist(err)` 区分"文件不存在"和"其他错误"。

---

### 坕 5：按行读取忽略最后一行无换行符

如果最后一行没有换行符，`scanner.Scan()` 仍然会返回该行。

不要假设每行都有换行符。

---

## 关键概念对比

| 函数 | 用途 | 返回值 | 是否自动关闭 |
|------|------|--------|--------------|
| `os.Open` | 打开文件只读 | `*os.File` | 需手动关闭 |
| `os.Create` | 创建/清空文件 | `*os.File` | 需手动关闭 |
| `os.ReadFile` | 一次性读取 | `[]byte` | 自动关闭 |
| `os.WriteFile` | 一次性写入 | `error` | 自动关闭 |
| `bufio.Scanner` | 按行读取 | 每行 `string` | 需关闭底层文件 |
| `bufio.Writer` | 缓冲写入 | 需 `Flush` | 需关闭底层文件 |

简单场景用 `os.ReadFile`/`os.WriteFile`；需要逐行或缓冲时用 `bufio`。