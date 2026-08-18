# Day 19 教案：middleware

## 学习目标

学完今天，需要能够做到：

1. 理解 middleware 在 Gin 请求链里的作用。
2. 编写 logging middleware。
3. 编写 CORS middleware。
4. 编写 request ID middleware。
5. 了解 auth placeholder 的用途和边界。
6. 能通过测试和 `curl` 验证 middleware 行为。

---

## Day 19 的位置

Day 18 已经把 repository 和数据库层打通，Day 19 开始把请求链路前面的横切能力补齐。

这一阶段先搭 middleware 骨架，后面再接 JWT、用户登录和权限控制。

---

## middleware 是什么

middleware 是位于请求处理链中间的一层。

它可以在 handler 之前做预处理，也可以在 handler 之后补充记录、统计、清理工作。

在 Gin 中，middleware 的典型写法是：

```go
func Something() gin.HandlerFunc {
    return func(c *gin.Context) {
        // before
        c.Next()
        // after
    }
}
```

---

## 执行顺序

Gin 会按注册顺序执行 middleware。

如果写成：

```go
router.Use(A(), B(), C())
```

执行顺序是：

1. A before
2. B before
3. C before
4. handler
5. C after
6. B after
7. A after

这也是为什么 logging middleware 通常把耗时统计放在 `c.Next()` 之后。

---

## Logging middleware

logging middleware 的目标是记录一次请求的基本信息：

- method
- path
- status
- latency

示例：

```go
func Logging() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        log.Printf("method=%s path=%s status=%d latency=%s", c.Request.Method, c.Request.URL.RequestURI(), c.Writer.Status(), time.Since(start))
    }
}
```

要点：

- `c.Next()` 必须先调用，才能拿到 handler 处理后的最终状态码。
- 日志只做记录，不改变响应内容。

---

## CORS middleware

CORS 用于浏览器跨域访问。

常见要处理的响应头：

- `Access-Control-Allow-Origin`
- `Access-Control-Allow-Methods`
- `Access-Control-Allow-Headers`
- `Access-Control-Expose-Headers`

浏览器发起跨域请求前，可能会先发 `OPTIONS` 预检请求。

这种请求通常不需要进入业务 handler，middleware 可以直接返回 `204 No Content`。

示例：

```go
if c.Request.Method == http.MethodOptions {
    c.AbortWithStatus(http.StatusNoContent)
    return
}
```

---

## Request ID middleware

request ID 用于链路追踪。

基本思路：

1. 如果请求头里已有 `X-Request-ID`，就直接使用。
2. 如果没有，就生成一个新的 ID。
3. 把它写回响应头。
4. 存到 Gin context 里，方便后面的 middleware 或 handler 读取。

示例：

```go
const RequestIDHeader = "X-Request-ID"

func RequestID() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := strings.TrimSpace(c.GetHeader(RequestIDHeader))
        if requestID == "" {
            requestID = newRequestID()
        }
        c.Set("request_id", requestID)
        c.Writer.Header().Set(RequestIDHeader, requestID)
        c.Next()
    }
}
```

这里推荐先把 ID 放进 context，而不是直接散落到全局变量里。

---

## Auth placeholder

Day 19 只需要预留 auth middleware 入口，不需要真正完成 JWT 验证。

原因：

- JWT 还没正式实现。
- 先把路由层的挂载点固定下来，后续接入更稳。
- 公开接口如 health 不应该被今天的占位逻辑拦住。

所以 placeholder 版本只需要透传请求即可。

---

## 今日代码结构

```text
internal/middleware/
  cors.go
  logging.go
  request_id.go
  auth.go
  middleware_test.go
cmd/server/main.go
```

---

## 今日验收标准

完成后应该满足：

1. `internal/middleware/` 中有可用的 Gin middleware。
2. `cmd/server/main.go` 挂载 middleware。
3. 至少写 3 组 middleware 测试。
4. `go test ./...` 通过。
5. 能通过 `curl` 验证响应头。
6. 能解释 middleware 的执行顺序和用途。

---

## 可选挑战题：Request ID

为请求增加 `X-Request-ID`。

验证点：

- 客户端传入时保留原值。
- 客户端没传时自动生成。
- 响应头里能看到对应值。

---

## 今天最容易踩的坑

### 坑 1：logging 放在 `c.Next()` 前面

这样拿不到最终状态码和完整耗时。

### 坑 2：CORS 忘记处理 `OPTIONS`

浏览器预检会失败，前端跨域请求发不出去。

### 坑 3：request ID 只写进响应头，不写进 context

后续 middleware 和 handler 就无法方便复用。

### 坑 4：auth placeholder 现在就开始拦请求

Day 19 的目标不是完成 JWT，而是预留结构，所以当前应该保持透传。
