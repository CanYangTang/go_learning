# Day 12 教案：JSON 请求和响应处理

## 学习目标

学完今天，需要能够做到：

1. 理解 `encoding/json` 的 `Marshal` 和 `Unmarshal`。
2. 理解 struct tag `json:"name"` 的作用。
3. 使用 `json.NewEncoder` 和 `json.NewDecoder` 处理 HTTP 请求体和响应体。
4. 实现 JSON 请求解析和 JSON 响应写入。
5. 编写 JSON handler 的测试。
6. 理解请求体读取和 `r.Body` 的生命周期。

---

## Day 12 的位置

Day 11 学了 HTTP server、route、handler 和 `/health` 健康检查。

Day 12 会把重点放在 JSON 数据的请求解析和响应写入，为后续 TODO API 做准备。

Day 13 会进入文件操作，Day 14 会做 HTTP 并发综合练习。

---

## JSON 是什么

JSON (JavaScript Object Notation) 是一种轻量级数据交换格式。

常见结构：

```json
{
  "name": "Alice",
  "age": 30,
  "active": true
}
```

在 Web API 中，JSON 是最常用的请求和响应格式。

---

## Go 与 JSON

Go 标准库 `encoding/json` 提供了 JSON 编解码能力。

常见操作：

- `json.Marshal`：Go 数据 → JSON 字节切片。
- `json.Unmarshal`：JSON 字节切片 → Go 数据。
- `json.NewEncoder(w).Encode(v)`：Go 数据 → JSON → 写入 `io.Writer`。
- `json.NewDecoder(r).Decode(&v)`：从 `io.Reader` 读取 JSON → Go 数据。

---

## struct tag `json:"name"`

Go struct 字段可以通过 tag 指定 JSON 字段名。

示例：

```go
type User struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}
```

- 序列化时，Go 字段 `Name` 会变成 JSON 字段 `"name"`。
- 反序列化时，JSON 字段 `"name"` 会映射到 Go 字段 `Name`。

如果不写 tag，默认使用字段名的大写形式作为 JSON 字段名。

常见 tag 选项：

- `json:"name"`：字段名映射。
- `json:"name,omitempty"`：零值时省略该字段。
- `json:"-"`：忽略该字段，不参与 JSON 序列化。

---

## `json.Marshal` 和 `json.Unmarshal`

### Marshal

```go
user := User{Name: "Alice", Age: 30}
data, err := json.Marshal(user)
if err != nil {
    // 处理错误
}
fmt.Println(string(data))
// {"name":"Alice","age":30}
```

返回的是字节切片 `[]byte`。

### Unmarshal

```go
data := []byte(`{"name":"Bob","age":25}`)
var user User
if err := json.Unmarshal(data, &user); err != nil {
    // 处理错误
}
fmt.Println(user.Name, user.Age)
// Bob 25
```

注意 `Unmarshal` 需要传指针 `&user`。

---

## `json.Encoder` 和 `json.Decoder`

`Marshal` 和 `Unmarshal` 操作的是字节切片。

`Encoder` 和 `Decoder` 操作的是 `io.Writer` 和 `io.Reader`。

在 HTTP handler 中，`http.ResponseWriter` 实现了 `io.Writer`，`r.Body` 实现了 `io.Reader`。

### Encoder

```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(response)
```

`Encode` 会把 Go 数据编码成 JSON 并写入 `w`。

### Decoder

```go
var req Request
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    // 处理错误
}
```

`Decode` 会从 `r.Body` 读取 JSON 并解码到 `req`。

---

## 请求体 `r.Body`

`r.Body` 是 `io.ReadCloser`，表示请求体。

读取后通常需要关闭：

```go
defer r.Body.Close()
```

但在 `json.NewDecoder(r.Body).Decode(&req)` 中，Decoder 会读取数据，不会主动关闭。

建议显式 `defer r.Body.Close()`，虽然某些框架或中间件会处理，但显式关闭是好习惯。

---

## JSON 请求处理流程

典型流程：

1. 解析请求体 JSON 到 struct。
2. 验证字段。
3. 处理业务逻辑。
4. 返回 JSON 响应。

示例：

```go
type CreateTodoRequest struct {
    Title string `json:"title"`
    Done  bool   `json:"done"`
}

type CreateTodoResponse struct {
    ID    int    `json:"id"`
    Title string `json:"title"`
    Done  bool   `json:"done"`
}

func createTodoHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    var req CreateTodoRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "invalid json"})
        return
    }
    defer r.Body.Close()

    // TODO: 业务逻辑，比如保存到数据库
    resp := CreateTodoResponse{
        ID:    1,
        Title: req.Title,
        Done:  req.Done,
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(resp)
}
```

---

## JSON 响应封装

Day 11 中 `pkg/response/response.go` 已经有 `JSON` 和 `Error` helper。

可以继续使用：

```go
response.JSON(w, http.StatusOK, todo)
response.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid json body")
```

今天可以在此基础上增加 JSON 请求解析的练习。

---

## 测试 JSON handler

测试 JSON handler 时，需要构造 JSON 请求体。

示例：

```go
body := `{"title":"Learn Go","done":false}`
req := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(body))
req.Header.Set("Content-Type", "application/json")
recorder := httptest.NewRecorder()

createTodoHandler(recorder, req)

if recorder.Code != http.StatusOK {
    t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
}

var resp CreateTodoResponse
if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
    t.Fatalf("decode response error = %v", err)
}

if resp.Title != "Learn Go" {
    t.Fatalf("title = %q, want %q", resp.Title, "Learn Go")
}
```

这里用 `strings.NewReader(body)` 创建请求体 reader。

---

## 今日代码练习设计

今天建议创建：

```text
week02-core/day12-json/
├── handler.go
└── handler_test.go
```

实现：

```go
type EchoRequest struct {
    Message string `json:"message"`
}

type EchoResponse struct {
    Message string `json:"message"`
}

func EchoHandler(w http.ResponseWriter, r *http.Request)
```

`EchoHandler` 行为：

1. 只接受 POST 方法。
2. 解析 JSON 请求体。
3. 返回相同的 message 作为响应。
4. 如果 JSON 解析失败，返回 400 和错误信息。

---

## 今日验收标准

完成后应该满足：

1. 创建 `week02-core/day12-json/`。
2. 实现 `EchoHandler`，能正确解析 JSON 请求并返回 JSON 响应。
3. 至少写 3 组测试：成功响应、无效 JSON、非 POST 方法。
4. `go test ./week02-core/day12-json` 通过。
5. `make test` 通过。
6. 能解释 `Marshal`、`Unmarshal`、struct tag、`Encoder`、`Decoder`、`r.Body`。

---

## 可选挑战题：简单验证

Issue 的可选挑战是增加基础验证。

例如：

- `message` 不能为空。
- `message` 长度不能超过 100。

如果验证失败，返回 400 和错误信息。

---

## 今天最容易踩的坑

### 坑 1：忘记传指针给 `Unmarshal` 或 `Decode`

错误：

```go
var req Request
json.Unmarshal(data, req)  // 错误，应该传 &req
```

正确：

```go
var req Request
json.Unmarshal(data, &req) // 正确
```

---

### 坑 2：JSON 字段名和 struct tag 不匹配

如果 JSON 是 `{"title":"Learn Go"}`，但 struct tag 是 `json:"name"`，字段会解析不到。

需要确保 tag 和 JSON 字段名一致。

---

### 坑 3：忘记设置 `Content-Type`

返回 JSON 时，应该设置：

```go
w.Header().Set("Content-Type", "application/json")
```

否则客户端可能按默认方式解析，导致问题。

---

### 坑 4：测试时忘记设置请求体或 Header

测试 JSON handler 时，需要：

```go
req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(body))
req.Header.Set("Content-Type", "application/json")
```

否则 handler 可能无法正确解析。

---

### 坑 5：未关闭 `r.Body`

虽然在简单场景下不关闭也能工作，但养成 `defer r.Body.Close()` 习惯更好。

---

## 关键概念对比

| 操作 | 输入 | 输出 | 用途 |
|------|------|------|------|
| `json.Marshal` | Go 数据 | `[]byte` | 内存中编码 |
| `json.Unmarshal` | `[]byte`, Go 指针 | 填充 Go 数据 | 内存中解码 |
| `json.NewEncoder(w).Encode(v)` | Go 数据, `io.Writer` | 写入 w | 流式编码 |
| `json.NewDecoder(r).Decode(&v)` | `io.Reader`, Go 指针 | 填充 Go 数据 | 流式解码 |

HTTP handler 中优先用 `Encoder` 和 `Decoder`，因为可以直接操作 `ResponseWriter` 和 `r.Body`。
