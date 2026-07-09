# Day 15 学习记录

## 日期

2026-07-09

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/15

## 今日教案

- 教案文档：`docs/daily/day-15-lesson.md`

## 核心任务

- 学习 Gin 框架的基本用法。
- 使用 `gin.New()` 创建路由器。
- 使用 `router.Group()` 创建 `/api/v1` 路由组。
- 实现 Gin handler 并返回 JSON 响应。
- 理解项目分层：`cmd/server` 和 `internal/handler` 的职责。

## 验收标准

- Gin 依赖已添加到 `go.mod`。
- `internal/handler/health.go` 实现 `HealthHandler`。
- `cmd/server/main.go` 使用 Gin 并创建 `/api/v1` 路由组。
- 测试覆盖 handler 成功响应和方法不允许。
- `go test ./internal/handler` 通过。
- `make test` 通过。
- `make run` + `curl http://localhost:8080/api/v1/health` 返回 JSON。
- 能解释 Gin 路由、路由组、handler、`gin.Context`。

## 可选挑战题

- 增加统一 404 JSON 响应。

## 答疑记录

- 待记录。

## 今日产出

- 添加 Gin 框架依赖到 `go.mod`。
- 实现 `internal/handler/health.go`：Gin health handler，返回 JSON 状态。
- 更新 `cmd/server/main.go`：使用 Gin 路由器，创建 `/api/v1` 路由组。
- 编写测试覆盖 handler 成功响应和方法不允许。

## 运行过的命令

```bash
go get -u github.com/gin-gonic/gin
go test ./internal/handler
make fmt
make test
make vet
```

## 代码 Review 结论

- `gin.SetMode(gin.ReleaseMode)` 设置生产模式，减少日志输出。
- `router := gin.New()` 创建空路由器，不附加默认中间件。
- `v1 := router.Group("/api/v1")` 创建 API v1 路由组，所有 `/api/v1/*` 路由归入此组。
- `v1.GET("/health", handler.HealthHandler)` 注册 health handler 到路由组。
- `HealthHandler` 使用 `c.JSON(http.StatusOK, gin.H{...})` 返回 JSON 响应，自动设置 Content-Type。
- `gin.SetMode(gin.TestMode)` 在测试中禁用日志输出。
- 测试使用 `router.ServeHTTP(recorder, req)` 模拟 HTTP 请求。
- `go test ./internal/handler`、`make fmt`、`make test`、`make vet` 均已通过。

## 今日小测试

1. Gin 框架的 `gin.New()` 和 `gin.Default()` 有什么区别？
   - 回答：`gin.New()` 创建一个空的 Gin 路由器。`gin.Default()` 会自动附加 Logger 和 Recovery 中间件。
   - 结果：通过。
   - 标准答案：`gin.New()` 创建空路由器，不附加任何中间件；`gin.Default()` 创建路由器并自动附加 Logger（日志）和 Recovery（panic 恢复）中间件。
2. `router.Group("/api/v1")` 创建的是什么？有什么好处？
   - 回答：创建一个路由组，组内所有路由自动加上 `/api/v1` 前缀。好处：避免重复写前缀、可以给整组路由统一加中间件（比如鉴权）、结构清晰，便于版本管理。
   - 结果：通过。
   - 标准答案：创建路由组，组内路由共享 `/api/v1` 前缀。好处：避免重复写前缀、统一添加中间件、便于 API 版本管理。
3. `gin.Context` 是什么？`c.JSON()` 做了什么？
   - 回答：`gin.Context` 是 Gin 对一次 HTTP 请求的完整封装，包含请求信息和响应方法。`c.JSON()` 做了三件事：设置 Content-Type: application/json、设置状态码、把数据序列化成 JSON 写入响应体。
   - 结果：通过。
   - 标准答案：`gin.Context` 封装了请求和响应，提供请求参数获取、响应写入等方法。`c.JSON(code, obj)` 设置状态码、设置 `Content-Type: application/json`，并将 obj 序列化为 JSON 写入响应体。
4. `gin.H` 是什么类型？为什么用 `gin.H{"status": "ok"}` 比较方便？
   - 回答：`type H map[string]any`。方便在返回简单 JSON 时不用定义结构体。
   - 结果：通过。
   - 标准答案：`gin.H` 是 `map[string]any` 的类型别名，用于快速构造 JSON 对象，无需定义结构体。
5. 为什么测试时要用 `gin.SetMode(gin.TestMode)`？
   - 回答：测试时不设置会：控制台输出大量路由注册日志，干扰测试输出影响测试可读性。设置 TestMode 让测试输出干净，只看到测试结果。
   - 结果：通过。
   - 标准答案：不设置 TestMode 时，Gin 会输出大量路由注册日志，干扰测试输出。设置 `gin.TestMode` 后禁用日志，让测试输出更干净。

## 测试结果

- 5 题全部通过。

## 遇到的问题

- 无。

## 关键收获

1. Gin 是对 `net/http` 的封装，提供路由分组、中间件、JSON 响应等开箱即用功能。
2. 路由组可以把一组路由放在同一个前缀下，便于 API 版本控制和中间件管理。
3. `gin.Context` 封装了请求和响应，`c.JSON()` 自动设置 Content-Type 并序列化 JSON。
4. `gin.H` 是 `map[string]any` 的简写，方便构造 JSON 响应。
5. 项目分层：`cmd/server` 负责组装和启动，`internal/handler` 负责处理 HTTP 请求。

## 明日计划

- 进入 Day 16：TODO 请求绑定和 JSON 响应。
