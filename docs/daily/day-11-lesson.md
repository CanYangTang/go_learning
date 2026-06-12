# Day 11 教案：HTTP Server 和 Health Endpoint

## 学习目标

学完今天，需要能够做到：

1. 理解 Go 标准库 `net/http` 的基础用法。
2. 理解 HTTP server、route、handler 的关系。
3. 使用 `http.HandleFunc` 注册路由。
4. 使用 `http.ListenAndServe` 启动服务。
5. 实现 `/health` 健康检查接口。
6. 理解 `http.ResponseWriter` 和 `*http.Request` 的作用。
7. 编写基础 HTTP handler 测试。
8. 能用 `make run` 和 `curl` 验证服务。

---

## Day 11 的位置

前面几天已经学习了：

- Day 8：goroutine 和 `sync.WaitGroup`。
- Day 9：channel 通信。
- Day 10：并发任务队列。

Day 11 开始进入 Web 后端主线。

今天先不引入 Gin，而是使用 Go 标准库 `net/http`，理解 HTTP server 的最小工作模型。

后面学习 Gin、middleware、handler/service/repository 分层时，今天的概念会继续使用。

---

## HTTP server 是什么

HTTP server 是一个长期运行的程序，它监听某个端口，等待客户端请求。

例如：

```text
浏览器 / curl / 前端应用
        ↓ HTTP request
Go HTTP server
        ↓ route 匹配
handler 处理请求
        ↓ HTTP response
客户端收到响应
```

在 Go 中，可以用标准库 `net/http` 启动 HTTP server。

---

## `net/http` 是什么

`net/http` 是 Go 标准库中的 HTTP 包。

它提供了：

- HTTP server 能力。
- HTTP client 能力。
- 路由注册函数。
- 请求和响应对象。
- 测试 HTTP handler 的工具。

今天重点使用 server 侧能力。

常见导入：

```go
import "net/http"
```

---

## route 和 handler

### route

route 是请求路径和处理逻辑之间的映射。

例如：

```text
GET /health -> healthHandler
```

当客户端访问 `/health` 时，server 会调用对应 handler。

### handler

handler 是真正处理 HTTP 请求的函数。

在 `net/http` 中，最常见的 handler 函数签名是：

```go
func(w http.ResponseWriter, r *http.Request)
```

例如：

```go
func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
}
```

---

## `http.ResponseWriter`

`http.ResponseWriter` 用来写 HTTP response。

常见能力：

1. 设置响应头。
2. 设置状态码。
3. 写响应体。

示例：

```go
w.Header().Set("Content-Type", "text/plain")
w.WriteHeader(http.StatusOK)
w.Write([]byte("ok"))
```

注意：

- `WriteHeader` 用来设置状态码。
- 如果没有显式调用 `WriteHeader`，第一次 `Write` 时默认会发送 `200 OK`。
- 状态码通常只应该写一次。

---

## `*http.Request`

`*http.Request` 表示客户端发来的请求。

常用信息包括：

- `r.Method`：HTTP 方法，例如 `GET`、`POST`。
- `r.URL.Path`：请求路径。
- `r.Header`：请求头。
- `r.Body`：请求体。

今天可以先读取 `r.Method`，限制 `/health` 只允许 `GET`。

---

## 注册路由：`http.HandleFunc`

`http.HandleFunc` 用来把路径和 handler 函数绑定起来。

示例：

```go
http.HandleFunc("/health", healthHandler)
```

意思是：当请求路径是 `/health` 时，交给 `healthHandler` 处理。

完整例子：

```go
func main() {
    http.HandleFunc("/health", healthHandler)
    http.ListenAndServe(":8080", nil)
}
```

第二个参数传 `nil` 时，表示使用默认路由器 `http.DefaultServeMux`。

---

## 启动服务：`http.ListenAndServe`

`http.ListenAndServe` 用来启动 HTTP server。

```go
err := http.ListenAndServe(":8080", nil)
```

含义：

- `:8080`：监听本机 8080 端口。
- `nil`：使用默认路由器。
- 返回值 `err`：如果 server 启动失败或异常退出，会返回错误。

通常写法：

```go
if err := http.ListenAndServe(":8080", nil); err != nil {
    log.Fatal(err)
}
```

`log.Fatal` 会打印错误并退出程序。

---

## Health endpoint 是什么

health endpoint 是健康检查接口。

常见路径：

```text
/health
/healthz
/ready
```

今天使用：

```text
GET /health
```

它通常用于：

- 判断服务是否启动成功。
- 给部署平台或监控系统检查服务状态。
- 本地开发时快速验证 server 是否能响应请求。

今天的目标响应可以很简单：

```text
status code: 200
body: ok
```

也可以返回 JSON：

```json
{"status":"ok"}
```

今天建议先使用 JSON，因为后续 TODO API 会大量返回 JSON。

---

## JSON 响应基础

Go 标准库可以用 `encoding/json` 编码 JSON。

例如：

```go
response := map[string]string{"status": "ok"}
json.NewEncoder(w).Encode(response)
```

写 JSON 前通常要设置响应头：

```go
w.Header().Set("Content-Type", "application/json")
```

完整 handler 示例：

```go
func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
```

---

## 限制 HTTP Method

如果 `/health` 只允许 GET，可以检查：

```go
if r.Method != http.MethodGet {
    w.WriteHeader(http.StatusMethodNotAllowed)
    return
}
```

`http.MethodGet` 是标准库提供的常量，值是 `"GET"`。

`http.StatusMethodNotAllowed` 是状态码 405。

这样可以避免 POST、PUT 等方法误用这个接口。

---

## Handler 测试

HTTP handler 可以不用真的启动端口也能测试。

Go 标准库提供 `net/http/httptest`。

常见测试结构：

```go
req := httptest.NewRequest(http.MethodGet, "/health", nil)
recorder := httptest.NewRecorder()

healthHandler(recorder, req)

if recorder.Code != http.StatusOK {
    t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
}
```

这里：

- `httptest.NewRequest` 创建一个假的 HTTP 请求。
- `httptest.NewRecorder` 记录 handler 写出的响应。
- 直接调用 handler，不需要启动真实 server。

---

## 今日代码练习设计

今天主要修改或新增：

```text
cmd/server/
├── main.go
└── main_test.go
```

建议实现：

```go
func healthHandler(w http.ResponseWriter, r *http.Request)
```

`main` 函数负责：

1. 注册 `/health` 路由。
2. 启动 `:8080` server。
3. 处理 `ListenAndServe` 返回的错误。

`healthHandler` 负责：

1. 只允许 GET。
2. 设置 `Content-Type: application/json`。
3. 返回 HTTP 200。
4. 返回 JSON：`{"status":"ok"}`。

---

## 为什么要把 handler 拆成函数

如果直接把逻辑写在 `main` 的匿名函数里，也能运行。

但拆成独立函数有几个好处：

- 更容易测试。
- `main` 更清晰，只负责组装和启动。
- 后续迁移到 handler package 或 Gin 时更容易重构。

今天先保持简单，不需要提前做复杂分层。

---

## 今日验收标准

完成后应该满足：

1. `cmd/server/` 中有基础 HTTP server。
2. `/health` 能返回成功响应。
3. `/health` 返回 JSON body。
4. 非 GET 方法返回 405。
5. 至少写 2 组 handler 测试。
6. `go test ./cmd/server` 通过。
7. `make test` 通过。
8. 能用 `make run` 启动服务。
9. 能用 `curl http://localhost:8080/health` 验证。
10. 能解释 HTTP server、route、handler、`ResponseWriter`、`Request`。

---

## 可选挑战题：简单请求日志

Issue 的可选挑战是增加简单 request logging。

可以在 handler 中临时打印：

```go
log.Printf("%s %s", r.Method, r.URL.Path)
```

但更推荐后续学 middleware 时再系统实现。

今天如果要做，可以保持非常简单，不要提前引入复杂抽象。

---

## 今天最容易踩的坑

### 坑 1：忘记处理 `ListenAndServe` 的错误

不建议这样写：

```go
http.ListenAndServe(":8080", nil)
```

更推荐：

```go
if err := http.ListenAndServe(":8080", nil); err != nil {
    log.Fatal(err)
}
```

---

### 坑 2：重复写状态码

一个 response 通常只写一次状态码。

如果先 `WriteHeader(200)`，后面再想写 `WriteHeader(500)`，通常已经来不及了。

---

### 坑 3：先写 body 再设置 header

响应头应该尽量在写 body 之前设置。

因为第一次 `Write` 可能会隐式发送状态码和响应头。

---

### 坑 4：测试时启动真实服务

handler 单元测试不需要真的监听端口。

优先使用 `httptest.NewRequest` 和 `httptest.NewRecorder`。

---

### 坑 5：默认路由器的全局状态

`http.HandleFunc` 使用默认全局路由器。

今天简单项目可以这样写。

但在更复杂测试中，全局路由器可能造成测试互相影响。

后续会学习使用独立的 `http.ServeMux` 或框架路由器。