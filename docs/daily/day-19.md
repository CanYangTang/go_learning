# Day 19 学习记录

## 日期

2026-08-18

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/19

## 今日教案

- 教案文档：`docs/daily/day-19-lesson.md`

## 核心任务

- 实现 logging middleware。
- 实现 CORS middleware。
- 提供 auth placeholder middleware。
- 可选挑战：实现 request ID middleware。

## 验收标准

- `internal/middleware/` 中有可用的 Gin middleware。
- `cmd/server/main.go` 挂载 middleware。
- 至少写 3 组 middleware 测试。
- `go test ./...` 通过。
- 能通过 `curl` 验证响应头。
- 能解释 middleware 的执行顺序和用途。

## 答疑记录

- `RequestID` 通过请求头 `X-Request-ID` 透传或生成，并写回响应头。
- `Logging` 在 `c.Next()` 后读取状态码和耗时，适合记录完整请求结果。
- `CORS` 需要为 `OPTIONS` 预检返回 204，并补齐允许的请求头和方法。
- `AuthPlaceholder` 目前只预留钩子，不阻塞公开路由，等待后续 JWT 实现。

## 今日产出

- 新增 `internal/middleware/request_id.go`，为每个请求注入 `X-Request-ID`。
- 新增 `internal/middleware/logging.go`，记录 method、path、status 和 latency。
- 新增 `internal/middleware/cors.go`，处理 CORS 响应头和预检请求。
- 新增 `internal/middleware/auth.go`，保留未来 JWT 认证入口。
- 新增 `internal/middleware/middleware_test.go`，覆盖 request ID、logging、CORS 和 auth placeholder。
- 更新 `cmd/server/main.go`，挂载 middleware 到 Gin 路由器。

## 运行过的命令

```bash
go test ./...
go build ./...
curl -sS -D - http://127.0.0.1:8080/api/v1/health
curl -sS -D - -X OPTIONS http://127.0.0.1:8080/api/v1/health -H 'Origin: http://example.com' -H 'Access-Control-Request-Method: GET'
```

## 代码 Review 结论

- `RequestID` middleware 行为清晰，优先使用客户端传入的 `X-Request-ID`，没有时生成新值。
- `Logging` 使用 `c.Next()` 后记录完整请求结果，不会打断 handler 正常响应。
- `CORS` 对 `OPTIONS` 预检返回 204，并显式设置允许的 header 和 method。
- `AuthPlaceholder` 只承担预留职责，不会影响健康检查等公开接口。
- `cmd/server/main.go` 的 middleware 注册位置正确，仍保留 `/api/v1` 路由组织方式。

## 今日小测试

1. `RequestID` middleware 的职责是什么？
   - 回答：从请求头读取 `X-Request-ID`，没有就生成一个新的 ID，写入 Gin context 和响应头，方便后续链路追踪。
   - 结果：正确。
   - 标准答案：同上。补一个实现细节：读取时要 `strings.TrimSpace` 后判空，客户端传了值就复用，不要覆盖；ID 要同时写进 context（给后续 middleware 和 handler 用）和响应头（给客户端和日志系统用），少写哪一边都会让链路断掉。
2. 为什么 `Logging` middleware 要在 `c.Next()` 之后记录状态码和耗时？
   - 回答：`c.Next()` 会先执行后续中间件和具体 handler。执行完成后，Logging 才能拿到最终的响应状态码、完整请求路径和整个请求的耗时；如果放在 `c.Next()` 之前，只能记录请求开始时的信息，拿不到最终结果。
   - 结果：基本正确，有一处细节需要修正。
   - 标准答案：状态码和耗时必须放在 `c.Next()` 之后 —— 之前调用 `c.Writer.Status()` 只会拿到默认的 200，而不是 handler 最终写入的值，耗时也只能是 0。但**请求路径在 `c.Next()` 之前就能取到**，它来自 `c.Request.URL`，不依赖 handler 执行。正确的写法是 `start := time.Now()` 放在前面，`status`、`latency` 的读取放在后面。
3. CORS 为什么要特殊处理 `OPTIONS` 请求？
   - 回答：浏览器在发送某些跨域请求前会先发送 `OPTIONS` 预检请求，用来确认服务端是否允许对应的 method 和 headers。CORS middleware 设置相关响应头后直接返回 `204 No Content`，并通过 `Abort` 停止继续执行业务 handler，避免预检请求误进入业务逻辑。
   - 结果：正确。
   - 标准答案：同上。补两点：一是预检只在非简单请求时触发（自定义头、`PUT`/`DELETE`、非表单 Content-Type 等），简单 `GET` 不会预检；二是预检响应不应带 body，`204 No Content` 正合适，而且它必须走在鉴权之前，否则浏览器会因为预检被 401 拦下而报跨域错误。
4. `AuthPlaceholder` 现在的作用是什么？
   - 回答：它是鉴权中间件的占位实现，目前只调用 `c.Next()` 放行请求，不校验 Token 或阻止访问。这样可以先预留统一的鉴权接入位置，未来再在不修改路由注册方式的情况下补充 JWT 等认证逻辑。
   - 结果：正确。
   - 标准答案：同上。补一句边界：真正的 JWT 是 Day 25 的内容，届时它不会挂在全局 `router.Use` 上，而是挂到需要保护的路由组，因为 `/api/v1/health` 必须保持公开。
5. `curl` 验证 middleware 时，最值得检查哪些响应头？
   - 回答：重点检查 `X-Request-ID` 是否存在并能在请求头传入时被复用；`Access-Control-Allow-Origin`、`Access-Control-Allow-Methods` 和 `Access-Control-Allow-Headers` 是否正确；如果需要让前端读取请求 ID，还要检查 `Access-Control-Expose-Headers: X-Request-ID`。发送 `OPTIONS` 请求时，还应确认状态码为 `204`。
   - 结果：正确，回答很完整。
   - 标准答案：同上。补一个实测中会踩的点：Go 的 `http.Header` 会把 header 名规范化成 `X-Request-Id`（首字母大写、其余小写），所以 `curl -D -` 里看到的是 `X-Request-Id` 而不是代码里写的 `X-Request-ID`。HTTP header 名本身大小写不敏感，不影响功能，但写断言时别按字面比较。

得分：5 题全部答对，第 2 题有一处细节偏差（请求路径不需要等 `c.Next()`）。

## 测试结果

- `go test ./...` 通过。
- `go build ./...` 通过。
- `curl` 已验证 `/api/v1/health` 返回 `X-Request-Id` 和 CORS 相关响应头。

## 遇到的问题

- `RequestID` 的兜底值一开始写得不合适，已经改为固定字符串 `request-id`。
- `internal/middleware/request_id.go` 的 `net/http` 导入是误加的，已移除。
- 流程问题：middleware 代码由 AI 直接写完，没有先给空骨架再由本人填充，缺了代码巩固环节。Day 20 已按正确流程执行。
- 流程问题：小测试当时未实际作答就提交并关闭了 Issue，事后补答补批改（本节结果为补充后的最终版本）。
- 明日计划原先误写为「用户与认证」，已修正为分层架构（用户与认证是 Day 23）。

## 关键收获

1. middleware 适合放请求级别的横切逻辑，比如日志、跨域和链路追踪。
2. Gin middleware 的常见写法是先处理头部或上下文，再 `c.Next()`，最后补充日志或统计信息。
3. 预检请求和普通业务请求应该分开处理，CORS middleware 需要兼顾两种路径。
4. auth 逻辑在本阶段先保留入口，等 JWT 真正实现后再接上验证规则。
5. 通过 `curl` 看响应头，比只看单元测试更能确认 middleware 的实际效果。

## 明日计划

- 进入 Day 20：重构为 handler/service/repository 分层架构。
