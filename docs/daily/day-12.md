# Day 12 学习记录

## 日期

2026-07-02

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/12

## 今日教案

- 教案文档：`docs/daily/day-12-lesson.md`

## 核心任务

- 学习 `encoding/json` 的 `Marshal` 和 `Unmarshal`。
- 理解 struct tag `json:"name"` 的作用。
- 使用 `json.NewEncoder` 和 `json.NewDecoder` 处理 HTTP 请求体和响应体。
- 实现 JSON 请求解析和 JSON 响应写入的 handler。

## 验收标准

- 创建 `week02-core/day12-json/`。
- 实现 `EchoHandler`，能正确解析 JSON 请求并返回 JSON 响应。
- 至少写 3 组测试：成功响应、无效 JSON、非 POST 方法。
- `go test ./week02-core/day12-json` 通过。
- `make test` 通过。
- 能解释 `Marshal`、`Unmarshal`、struct tag、`Encoder`、`Decoder`、`r.Body`。

## 可选挑战题

- 增加基础验证：`message` 不能为空，长度不能超过 100。

## 答疑记录

- 待记录。

## 今日产出

- 创建 `week02-core/day12-json/` 目录。
- 实现 `EchoRequest`、`EchoResponse`、`ErrorResponse` 结构体，带 `json` tag。
- 实现 `EchoHandler`：只接受 POST，解析 JSON 请求体，返回相同 message。
- 编写测试覆盖成功响应、无效 JSON 返回 400、非 POST 返回 405。

## 运行过的命令

```bash
go test ./week02-core/day12-json
make fmt
make test
make vet
```

## 代码 Review 结论

- `EchoHandler` 在开头检查 `r.Method != http.MethodPost`，非 POST 直接返回 405，结构清晰。
- 使用 `json.NewDecoder(r.Body).Decode(&req)` 从请求体流式解码 JSON，适合 HTTP 场景。
- JSON 解析失败时返回 400 和 `{"error":"..."}`，错误信息硬编码为 `"..."`。
- `defer r.Body.Close()` 在解码成功后关闭请求体，位置正确。
- 成功时设置 `Content-Type: application/json`，返回 200 和 JSON 响应。
- `var res = req` 利用 EchoRequest 和 EchoResponse 字段相同，直接复用 struct 进行编码。
- 测试使用 `strings.NewReader(body)` 构造 JSON 请求体，用 `httptest` 验证 handler 行为。
- `go test ./week02-core/day12-json`、`make fmt`、`make test`、`make vet` 均已通过。

## 今日小测试

1. `json.Marshal` 和 `json.Unmarshal` 分别做什么？输入和输出是什么？
   - 回答：Go 数据 → JSON 字节切片，JSON 字节切片 → Go 数据。
   - 结果：通过。
   - 标准答案：`json.Marshal` 把 Go 数据编码成 JSON `[]byte`；`json.Unmarshal` 把 JSON `[]byte` 解码到 Go struct（需要传指针）。
2. struct tag `json:"message"` 有什么作用？如果不写 tag，JSON 字段名会是什么？
   - 回答：作用：控制 JSON 序列化/反序列化时的字段名。如果不写 tag，JSON 字段名就是 Go 字段名本身。
   - 结果：通过。
   - 标准答案：`json:"message"` 指定 Go 字段与 JSON 字段 `"message"` 的映射关系。不写 tag 时，默认使用 Go 字段名（大写形式）作为 JSON 字段名。
3. `json.NewEncoder(w).Encode(v)` 和 `json.Marshal(v)` 有什么区别？什么时候用 Encoder？
   - 回答：Encoder 直接写入 `io.Writer`，Marshal 返回 `[]byte`，数据在内存。写 HTTP 响应、写文件，目标是一个 `io.Writer` 时用 Encoder。
   - 结果：通过。
   - 标准答案：`Encoder.Encode(v)` 直接把 JSON 写入 `io.Writer`，适合 HTTP response body、文件等流式场景；`Marshal(v)` 返回 `[]byte`，适合需要先拿到完整 JSON 字节再处理的场景。
4. `json.NewDecoder(r.Body).Decode(&req)` 为什么能直接从 HTTP 请求体解码？`r.Body` 是什么类型？
   - 回答：`r.Body` 的类型是 `io.ReadCloser`。`json.NewDecoder` 接受 `io.Reader`，而 `r.Body` 满足这个接口，所以可以直接传进去。`NewDecoder` 会一边从 body 读字节，一边解析 JSON，不需要先把 body 全部读到内存再处理。
   - 结果：通过。
   - 标准答案：`r.Body` 是 `io.ReadCloser`，实现了 `io.Reader` 接口。`json.NewDecoder` 接受 `io.Reader`，所以可以直接用 `r.Body`。Decoder 会流式读取并解析，比先读到内存再 Unmarshal 更高效。
5. 为什么 `defer r.Body.Close()` 要放在函数开头，而不是放在解码成功后？
   - 回答：`defer` 在函数返回时才执行，不是立刻执行。所以放开头只是"提前登记"，实际关闭时机仍在函数末尾，不影响中间读取 body。
   - 结果：基本通过，需要补充关键原因。
   - 标准答案：`defer` 放在开头是为了确保所有返回路径（包括方法检查失败、JSON 解码失败等 early return）都会执行关闭。如果放在解码成功后，解码失败时直接 `return` 就不会关闭 `r.Body`，可能导致资源泄漏。

## 测试结果

- 4 题通过，1 题基本通过。
- 需要补充：`defer r.Body.Close()` 放开头的关键原因是覆盖所有 early return 路径，防止资源泄漏。

## 遇到的问题

- 无。

## 关键收获

1. `json.Marshal` 把 Go 数据编码成 JSON 字节切片，`json.Unmarshal` 把 JSON 字节切片解码到 Go struct。
2. struct tag `json:"name"` 指定 JSON 字段名映射；不写 tag 时默认使用 Go 字段名的大写形式。
3. `Encoder/Decoder` 可以直接操作 `io.Writer` 和 `io.Reader`，适合 HTTP handler 场景。
4. `r.Body` 是 `io.ReadCloser`，表示 HTTP 请求体，读取后需要关闭。
5. `defer r.Body.Close()` 放在解码成功后可以确保正常流程关闭请求体。

## 明日计划

- 进入 Day 13：文件操作练习。
