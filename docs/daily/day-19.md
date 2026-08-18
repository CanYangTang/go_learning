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
   - 结果：待批改。
2. 为什么 `Logging` middleware 要在 `c.Next()` 之后记录状态码和耗时？
   - 回答：待批改。
3. CORS 为什么要特殊处理 `OPTIONS` 请求？
   - 回答：待批改。
4. `AuthPlaceholder` 现在的作用是什么？
   - 回答：待批改。
5. `curl` 验证 middleware 时，最值得检查哪些响应头？
   - 回答：待批改。

## 测试结果

- `go test ./...` 通过。
- `go build ./...` 通过。
- `curl` 已验证 `/api/v1/health` 返回 `X-Request-Id` 和 CORS 相关响应头。

## 遇到的问题

- `RequestID` 的兜底值一开始写得不合适，已经改为固定字符串 `request-id`。
- `internal/middleware/request_id.go` 的 `net/http` 导入是误加的，已移除。

## 关键收获

1. middleware 适合放请求级别的横切逻辑，比如日志、跨域和链路追踪。
2. Gin middleware 的常见写法是先处理头部或上下文，再 `c.Next()`，最后补充日志或统计信息。
3. 预检请求和普通业务请求应该分开处理，CORS middleware 需要兼顾两种路径。
4. auth 逻辑在本阶段先保留入口，等 JWT 真正实现后再接上验证规则。
5. 通过 `curl` 看响应头，比只看单元测试更能确认 middleware 的实际效果。

## 明日计划

- 进入 Day 20：开始用户与认证相关内容。
