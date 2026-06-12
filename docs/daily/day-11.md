# Day 11 学习记录

## 日期

2026-06-12

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/11

## 今日教案

- 教案文档：`docs/daily/day-11-lesson.md`

## 核心任务

- 学习 Go 标准库 `net/http` 的基础 HTTP server 模型。
- 使用 route 和 handler 实现 `/health` 健康检查接口。
- 编写基础 handler 测试，并用 `make run` + `curl` 验证服务。

## 验收标准

- `cmd/server/` 中有基础 HTTP server。
- `/health` 能返回成功响应。
- `/health` 返回 JSON body。
- 非 GET 方法返回 405。
- 至少写 2 组 handler 测试。
- `go test ./cmd/server` 通过。
- `make test` 通过。
- 能用 `make run` 启动服务。
- 能用 `curl http://localhost:8080/health` 验证。
- 能解释 HTTP server、route、handler、`ResponseWriter`、`Request`。

## 可选挑战题

- 增加简单 request logging。

## 答疑记录

- 默认路由器指 `net/http` 包中预先创建好的全局路由器 `http.DefaultServeMux`。
- `http.HandleFunc("/health", handler)` 会把路由注册到 `http.DefaultServeMux`。
- `http.ListenAndServe(":8080", nil)` 的第二个参数传 `nil` 时，也会使用 `http.DefaultServeMux` 来匹配请求路径。
- `ServeMux` 可以理解为请求分发器或路由表，维护路径和 handler 的映射关系，例如 `"/health" -> healthHandler`。
- 简单项目可以使用默认路由器；复杂项目和测试中更推荐使用 `http.NewServeMux()` 创建独立路由器，避免污染全局状态。

## 今日产出

- 使用标准库 `net/http` 实现基础 HTTP server。
- 使用 `http.NewServeMux()` 创建独立路由器，并注册 `GET /health`。
- `/health` 返回 JSON 响应，包含 `status`、`service`、`version`。
- 新增 `cmd/server/main_test.go`，覆盖 handler 成功响应和非 GET 方法返回 405。

## 运行过的命令

```bash
go test ./cmd/server
make fmt
make test
make vet
go list ./...
make run
curl http://localhost:8080/health
```

## 代码 Review 结论

- `main` 中使用 `http.NewServeMux()` 创建独立路由器，避免依赖默认全局路由器，结构更清晰。
- `mux.HandleFunc("GET /health", healthHandler)` 使用 Go 1.22+ 的 method-aware pattern，能让非 GET 请求自动返回 405。
- `http.ListenAndServe(addr, mux)` 显式传入路由器，server 启动逻辑清楚。
- `healthHandler` 通过 `response.JSON` 统一返回 JSON，并设置 HTTP 200。
- `response.writeJSON` 通过 `ResponseWriter` 设置响应头、状态码和 JSON body；`_ = json.NewEncoder(w).Encode(body)` 表示显式忽略编码错误。
- 测试使用 `httptest.NewRequest` 和 `httptest.NewRecorder` 直接测试 handler，不需要真实监听端口。
- 路由测试使用独立 `ServeMux` 验证 POST `/health` 返回 405。
- `go test ./cmd/server`、`make fmt`、`make test`、`make vet`、`make run` + `curl /health` 均已通过。

## 今日小测试

1. `http.NewServeMux()` 创建的是什么？它和 `http.DefaultServeMux` 有什么区别？
   - 回答：路由表；一个是自建的，一个是系统默认的。
   - 结果：通过。
   - 标准答案：`http.NewServeMux()` 创建一个新的独立 `ServeMux`，可以理解为独立路由器或路由表。`http.DefaultServeMux` 是 `net/http` 包提供的全局默认路由器；自建 `ServeMux` 不污染全局状态，测试和复杂项目中更可控。
2. `mux.HandleFunc("GET /health", healthHandler)` 中的 `GET /health` 表示什么？为什么 POST `/health` 会返回 405？
   - 回答：GET 方法访问 `/health` 路由；因为 POST 方法在 `healthHandler` 中不被允许。
   - 结果：基本通过，需要修正责任位置。
   - 标准答案：`GET /health` 是 Go 1.22+ `ServeMux` 的 method-aware pattern，表示只有 GET `/health` 会匹配到 `healthHandler`。POST `/health` 返回 405 是 `ServeMux` 根据路由模式自动处理的，不是 `healthHandler` 内部手动判断的。
3. `healthHandler(w http.ResponseWriter, r *http.Request)` 中 `w` 和 `r` 分别负责什么？
   - 回答：`w` 负责返回响应信息，`r` 负责读取请求信息。
   - 结果：通过。
   - 标准答案：`w` 是 `http.ResponseWriter`，用于写响应头、状态码和响应体；`r` 是 `*http.Request`，表示客户端请求，可以读取请求方法、路径、请求头、请求体等信息。
4. `httptest.NewRequest` 和 `httptest.NewRecorder` 在测试里分别起什么作用？为什么测试 handler 时不需要真的启动端口？
   - 回答：分别模拟请求和记录请求后的结果；为什么不需要真的启动端口还不清楚。
   - 结果：基本通过，需要补充原因。
   - 标准答案：`httptest.NewRequest` 创建一个测试用的 HTTP request；`httptest.NewRecorder` 创建一个记录 response 的对象，它实现了 handler 需要的 `ResponseWriter` 行为。测试时可以直接调用 `healthHandler(recorder, req)` 或 `mux.ServeHTTP(recorder, req)`，所以不需要真的启动 TCP 端口。
5. `json.NewEncoder(w).Encode(body)` 为什么算是在返回响应？这里的 `_ =` 表示什么？
   - 回答：因为 `w` 就是负责返回响应的；`_ =` 代表不处理返回的错误。
   - 结果：通过。
   - 标准答案：`w` 是当前请求的响应写入器，`Encode(body)` 会把 Go 数据编码成 JSON 并写入 `w`，也就是写入 HTTP response body。`_ =` 是空白标识符，表示显式丢弃 `Encode` 返回的 `error`。

## 测试结果

- 3 题通过，2 题基本通过。
- 需要修正：POST `/health` 返回 405 是 `ServeMux` 根据 `GET /health` 路由模式自动处理，不是 `healthHandler` 内部判断。
- 需要补充：handler 测试不需要启动端口，是因为可以直接用测试 request 和 recorder 调用 handler 或 `ServeMux.ServeHTTP`。

## 遇到的问题

- 对默认路由器和全局路由表的概念需要拆开理解：`DefaultServeMux` 是标准库提供的全局 `ServeMux`，而 `http.NewServeMux()` 会创建独立路由器。
- `ResponseWriter` 不只是一个普通变量，它代表当前请求要写回客户端的响应。
- `_ = json.NewEncoder(w).Encode(body)` 中的 `_` 是空白标识符，用来显式丢弃 `Encode` 返回的错误。

## 关键收获

1. HTTP server 接收请求后，会通过路由器匹配路径和方法，再调用对应 handler。
2. `ServeMux` 是 Go 标准库中的请求分发器，可以维护 route 到 handler 的映射关系。
3. `ResponseWriter` 用来写响应头、状态码和响应体；`*http.Request` 表示客户端发来的请求。
4. `httptest` 可以在不启动真实端口的情况下测试 handler 行为。
5. `json.NewEncoder(w).Encode(body)` 会把 Go 数据编码为 JSON 并写入 HTTP response body。

## 明日计划

- 进入 Day 12：JSON request/response handling。
