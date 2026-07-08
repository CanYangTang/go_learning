# Day 13 学习记录

## 日期

2026-07-08

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/13

## 今日教案

- 教案文档：`docs/daily/day-13-lesson.md`

## 核心任务

- 学习 `os` 包的文件打开、创建、读取、写入操作。
- 学习 `bufio` 包的缓冲读写和按行读取。
- 学习 `io` 包的 `Reader` 和 `Writer` 接口。
- 实现文件读取、写入、按行读取的函数。

## 验收标准

- 创建 `week02-core/day13-files/`。
- 实现 `ReadAll`、`ReadLines`、`WriteAll`、`WriteLines`。
- 至少写 4 组测试。
- `go test ./week02-core/day13-files` 通过。
- `make test` 通过。
- 能解释 `os.Open`、`os.Create`、`os.ReadFile`、`os.WriteFile`、`bufio.Scanner`、`bufio.Writer`、`io.Reader`、`io.Writer`。

## 可选挑战题

- 读取 JSON 文件并解析成 struct。

## 答疑记录

- struct 是数据结构，定义一组字段，存储具体数据。
- interface 是行为契约，定义一组方法签名，不存储数据。
- interface 不需要显式声明实现，只要类型的方法匹配接口的方法签名，就自动满足该接口。
- 例如 `*os.File` 是 struct，因为实现了 `Read` 方法，所以自动满足 `io.Reader` 接口。

## 今日产出

- 创建 `week02-core/day13-files/` 目录。
- 实现 `ReadAll`：用 `os.ReadFile` 一次性读取整个文件内容。
- 实现 `ReadLines`：用 `bufio.Scanner` 按行读取文件。
- 实现 `WriteAll`：用 `os.WriteFile` 一次性写入内容。
- 实现 `WriteLines`：用 `bufio.Writer` 按行写入，每行追加换行符。
- 编写测试覆盖成功读取、文件不存在、空文件、成功写入、空写入等场景。

## 运行过的命令

```bash
go test ./week02-core/day13-files
make fmt
make test
make vet
```

## 代码 Review 结论

- `ReadAll` 使用 `os.ReadFile` 一次性读取，简洁且自动关闭文件，适合小文件场景。
- `ReadLines` 使用 `bufio.Scanner` 按行读取，`scanner.Scan()` 循环配合 `scanner.Text()` 收集每行，适合逐行处理。
- `WriteAll` 使用 `os.WriteFile` 一次性写入，权限 `0644` 表示 owner 可读写、其他用户可读。
- `WriteLines` 使用 `bufio.NewWriter` 缓冲写入，每行调用 `WriteString(line + "\n")`，最后必须 `Flush()` 确保缓冲区内容写入文件。
- 所有文件操作都正确处理了错误返回，`defer file.Close()` 放在打开成功后确保资源释放。
- 测试使用 `os.CreateTemp` 创建临时文件，用 `defer os.Remove` 清理，避免污染工作目录。
- `go test ./week02-core/day13-files`、`make fmt`、`make test`、`make vet` 均已通过。

## 今日小测试

1. `os.ReadFile` 和 `os.Open` + 手动读取有什么区别？什么时候用 `ReadFile`？
   - 回答：`os.ReadFile` 是一次性读取全部内容，适合文件不大、需要完整内容，如读配置文件、模板等；`os.Open` + 手动读取是流式处理。
   - 结果：通过。
   - 标准答案：`os.ReadFile` 一次性读取整个文件并自动关闭，适合小文件；`os.Open` 返回 `*os.File`，需要手动读取和关闭，适合逐行处理或大文件流式读取。
2. `bufio.Scanner` 的 `Scan()` 和 `Text()` 分别做什么？循环什么时候结束？
   - 回答：`Scan()`：尝试读下一行，读到返回 `true`，到末尾或出错返回 `false`。`Text()`：返回 `Scan()` 刚读到的那行字符串。循环结束条件：文件读到末尾（EOF）或者读取出错（`scanner.Err()` 可以检查）。
   - 结果：通过。
   - 标准答案：`Scan()` 推进到下一行并返回 `true`，EOF 或出错时返回 `false`；`Text()` 返回当前行的内容（不含换行符）。`for scanner.Scan()` 循环在 `Scan()` 返回 `false`时结束。
3. `bufio.Writer` 为什么需要 `Flush()`？如果忘记 `Flush` 会发生什么？
   - 回答：`bufio.Writer` 有一个内部缓冲区，写数据时先攒在缓冲区里，不是每次都立刻写磁盘（减少系统调用，提升性能）。忘记 `Flush` 会怎样：缓冲区里的数据丢失，文件内容不完整或为空。
   - 结果：通过。
   - 标准答案：`bufio.Writer` 缓冲写入以减少系统调用次数。`Flush()` 把缓冲区内容写入底层 `io.Writer`。忘记 `Flush` 时，缓冲区数据可能不会写入文件，导致内容丢失或不完整。
4. `os.CreateTemp("", "test*.txt")` 创建的文件在哪里？为什么要用 `defer os.Remove(file.Name())`？
   - 回答：在系统临时目录。临时文件用完后需要手动删除，否则会残留在磁盘上。`defer` 确保函数退出时一定清理。
   - 结果：通过。
   - 标准答案：第一个参数 `""` 表示使用系统默认临时目录（如 `/tmp`）。`defer os.Remove(file.Name())` 确保测试结束后删除临时文件，避免残留占用磁盘空间。
5. `[]string` 和 `[3]string` 有什么区别？哪个能用 `append`？
   - 回答：一个是切片一个是数组，切片能用。
   - 结果：通过。
   - 标准答案：`[]string` 是切片，长度动态可变，可以用 `append` 增长；`[3]string` 是数组，长度固定为 3，不能用 `append`。

## 测试结果

- 5 题全部通过。

## 遇到的问题

- 初始化 slice 需要指定元素类型：`lines := []string{}`，不能写 `lines := []`。
- `append` 是函数，不是 slice 的方法，必须 `lines = append(lines, line)`。
- `WriteLines` 需要在每行后面加 `"\n"`，否则测试期望 `"line1\nline2\nline3\n"` 会失败。

## 关键收获

1. `os.ReadFile`/`os.WriteFile` 是一步读写，自动关闭文件，适合小文件；`os.Open`/`os.Create` 需要手动关闭。
2. `bufio.Scanner` 按行读取，`bufio.Writer` 缓冲写入，适合逐行处理或大文件场景。
3. `bufio.Writer` 必须调用 `Flush()`，否则缓冲区内容不会写入文件。
4. 测试文件操作时用 `os.CreateTemp` 创建临时文件，配合 `defer os.Remove` 清理。
5. `[]T` 是切片，长度可变，可用 `append`；`[N]T` 是数组，长度固定。

## 明日计划

- 进入 Day 14：HTTP 并发综合练习。
