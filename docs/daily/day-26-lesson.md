# Day 26 教案：统一错误 / 响应出口与分层收敛

> Issue #26：Refactor errors, responses, and layering。
> 核心：把错误处理、响应格式、分层边界收敛干净。
> **范围决策：只做重构**，配置迁环境变量（A4/B1/D5/E2）留独立一天。

## 0. 今天要动的东西（先看全景）

对照 `docs/issues-backlog.md`，今天清掉这几条：

- **B2** — `pkg/response` 里 `JSON`/`Error`/`GinJSON`/`GinError` 四个函数零调用，其中 `GinJSON`/`GinError` 是「参数 `c any` + 内联接口断言 + 断言失败静默返回」的反模式。删。
- **错误信封三处重复** — `handler.writeError`、`middleware/recovery.go`、`middleware/auth.go` 各自手拼一遍 `response.ErrorBody{Error: response.ErrorPayload{...}}`。收敛成 `pkg/response` 里唯一出口 `WriteError(c, err)`。
- **响应格式对称** — 成功路径 `c.JSON(status, response.Body{Data: x, Message: "ok"})` 在 handler 里重复 6 次，配一个对称的 `WriteSuccess(c, status, data)`。
- **C4** — `pkg/response` 没有测试，补上。
- **B4** — `internal/handler/health_test.go` 手写了 `containsString`/`findSubstring`（等于重造 `strings.Contains`），换标准库。
- **分层验证** — service/handler/auth 的依赖边界已经是干净的（下面有实测），今天补上编译期接口断言把它钉死。

## 1. 为什么删 GinJSON/GinError（不是「没人用」这么简单）

```go
// 反模式，今天删掉
func GinJSON(c any, statusCode int, data any) {
    type ginContext interface{ JSON(code int, obj any) }
    if gc, ok := c.(ginContext); ok {      // 断言失败……
        gc.JSON(statusCode, Body{Data: data, Message: "ok"})
    }
    // ……就什么都不做，静默返回，调用方以为写了响应，其实没写
}
```

两个问题叠加：

1. **绕开了编译期检查**。`c any` + 运行时断言，等于把「这里必须是 gin.Context」的约束从编译期推迟到运行时。真要依赖 gin 就直接写 `*gin.Context`，编译器帮你查。
2. **失败静默**。断言不过就 `return`，不 panic 不报错，响应体是空的，状态码是默认的 200。这种 bug 只会在运行时以「客户端收到空 200」的形式暴露。

它当初是为了「让 `pkg/response` 不依赖 gin」。但我们今天的结论正相反：**`pkg/response` 就是 HTTP 响应层，让它依赖 gin 是正确的**，这样才能给出 `WriteError(c *gin.Context, err error)` 这种带编译期检查的干净出口。

<!-- APPEND-HERE -->

## 2. 统一错误出口 `WriteError(c, err)`

### 2.1 为什么不会成环

收敛的前提是 `pkg/response` 可以同时 import `gin` 和 `apperror` 而不产生 import 环。实测确认（静态查）：

- import `pkg/response` 的只有 `internal/handler/*` 和 `internal/middleware/*`——它们是 `response` 的**下游**，`response` 不反过来 import 它们。
- `gin` 不 import 本项目任何包（`grep CanYangTang $GOMODCACHE/.../gin@*` 为空）。
- `apperror` 只 import `net/http`，是叶子包。

所以 `response → gin`、`response → apperror` 都是指向叶子/外部的边，不成环。这也解释了为什么之前要用 `c any` 的反模式绕依赖——那是**误判**，直接依赖 gin 本就安全。

### 2.2 一个出口，两种角色都对（实测）

`WriteError` 内部用 `c.AbortWithStatusJSON`。关键问题：handler 里用它、middleware 里也用它，行为都对吗？用一次性程序验证过（已删除）：

```
handler 角色: status=400 body={"error":{"code":"V","message":"bad"}} 后续 Next 处理 aborted=正确感知
middleware 角色: status=401 body={"error":{"code":"U","message":"no"}} 下游 handler reached=false
```

结论：

- **handler 角色**：写状态码 + 信封，同时标记 `Abort`。handler 本来就 `return`，Abort 无副作用；而 `Logging` 在 `c.Next()` 之后照样读 `c.Writer.Status()` 写访问日志（Recovery 早就用 `AbortWithStatusJSON` 验证过这条）。
- **middleware 角色**：写响应 + `Abort` 阻止下游 handler 执行——这正是 401 时需要的。

所以一个基于 `AbortWithStatusJSON` 的出口同时满足两种调用点，不需要「handler 用 `c.JSON`、middleware 用 `Abort`」两套。

### 2.3 出口的签名与语义

```go
// pkg/response/response.go
// WriteError 是错误响应的唯一出口。errors.As 命中 apperror.Error 就按它的
// StatusCode/Code/Message 写；命中不了（未归类的 error）一律 500 + 固定文案，
// 原始 error 绝不外泄（内部信息只应进日志，不进响应体）。
func WriteError(c *gin.Context, err error) {
    var appErr apperror.Error
    if errors.As(err, &appErr) {
        c.AbortWithStatusJSON(appErr.StatusCode, ErrorBody{
            Error: ErrorPayload{Code: appErr.Code, Message: appErr.Message},
        })
        return
    }
    c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorBody{
        Error: ErrorPayload{Code: "INTERNAL_ERROR", Message: "internal server error"},
    })
}
```

三处旧代码收敛后：

- `handler.writeError` **删除**，所有调用点直接 `response.WriteError(c, err)`。
- `middleware/auth.go` 的 `writeUnauthorized` **删除**，改 `response.WriteError(c, apperror.Unauthorized(msg))`。
- `middleware/recovery.go` 改 `response.WriteError(c, apperror.Internal("internal server error"))`。

### 2.4 对称的成功出口

```go
// WriteSuccess 是成功响应的唯一出口，固定 message="ok"。data 为 nil 时
// Body.Data 的 omitempty 会把它省掉（DELETE 成功就是 {"message":"ok"}）。
func WriteSuccess(c *gin.Context, statusCode int, data any) {
    c.JSON(statusCode, Body{Data: data, Message: "ok"})
}
```

成功路径不需要 `Abort`（handler 是链尾），用 `c.JSON` 即可。handler 里 6 处 `c.JSON(..., response.Body{...})` 收敛到它。

## 3. 分层边界（今天只验证 + 钉死，不重构）

实测（`grep` 确认）当前边界已经是干净的：

- `internal/service/` 不 import `gin` 也不 import `gorm`（纯业务）。
- `internal/handler/` 不 import `gorm`（不碰持久化细节）。
- `internal/auth/` 不 import `gin` 也不 import `gorm`。

今天做的是把这些「隐性约定」变成**编译期断言**，放在各自的 wiring 处（`main.go` 已有 `var _ TodoRepository = (*repository.TodoRepository)(nil)` 这类）。核对已有断言是否覆盖 `UserService`/`UserRepository`，缺的补上——这样接口漂移会在编译期而不是运行期爆。

## 4. C4 / B4 —— 补测试、删重造轮子

- **C4**：`pkg/response` 新增 `response_test.go`，覆盖 `WriteError`（apperror 命中→对应 status/code；非 apperror→500 固定文案；确认 `Abort` 生效）、`WriteSuccess`（status/body/`data=nil` 时省略 data）。用 `gin.SetMode(gin.TestMode)` + `httptest`。
- **B4**：`internal/handler/health_test.go` 删掉 `containsString`/`findSubstring`，`import "strings"` 用 `strings.Contains`。同包 `todo_test.go` 早就用标准库了，统一。

## 5. 今日代码结构（改动面）

```
pkg/response/response.go        删 JSON/Error/GinJSON/GinError/writeJSON；加 WriteError/WriteSuccess（依赖 gin+apperror）
pkg/response/response_test.go   新增（C4）
internal/handler/todo.go        删 writeError；调用点改 response.WriteError / response.WriteSuccess
internal/handler/user.go        调用点改 response.WriteError / response.WriteSuccess
internal/middleware/auth.go     删 writeUnauthorized；改 response.WriteError(c, apperror.Unauthorized(...))
internal/middleware/recovery.go 改 response.WriteError(c, apperror.Internal(...))
internal/handler/health_test.go 删 containsString/findSubstring，用 strings.Contains（B4）
cmd/server/main.go              补齐 UserService/UserRepository 编译期断言（若缺）
```

## 6. 分工（教案先行 + 空骨架由用户填充）

- **用户填**：`pkg/response` 的 `WriteError` 和 `WriteSuccess` 两个函数体（今天的核心逻辑）。
- **AI 做**：删 B2 死代码、把 handler/middleware 的调用点接到新出口（wiring）、删 `handler.writeError`/`writeUnauthorized`、B4 清理、编译期断言、所有测试文件。

骨架里 `WriteError`/`WriteSuccess` 是 `panic("not implemented")`，填之前全项目 build 得过（签名在）、但相关测试会 panic——这是骨架待填的正常状态，填完即绿。

## 7. 验收标准

1. `pkg/response` 不再有 `JSON`/`Error`/`GinJSON`/`GinError`/`writeJSON`，`grep` 零残留。
2. 错误信封只在 `pkg/response.WriteError` 一处构造；`handler.writeError`、`middleware.writeUnauthorized` 已删除，无其他地方手拼 `ErrorBody{...}`。
3. 成功响应统一走 `response.WriteSuccess`。
4. 未归类 error 仍返回 500 + 固定文案，原始 error 不外泄。
5. `pkg/response` 有测试，覆盖成功/apperror/非 apperror 三条路径。
6. `health_test.go` 用 `strings.Contains`，无手写子串查找。
7. 分层断言覆盖所有 consumer 接口。
8. `gofmt -l .` 干净、`go build ./...`、`go vet ./...`、`go test -count=1 ./internal/... ./pkg/...` 全绿；端到端行为不变（成功/400/401/404/500 信封与 Day 25 一致）。

## 8. 今天最容易踩的坑

1. **删函数留下未用 import**：删掉 `JSON`/`Error`/`writeJSON` 后，`encoding/json`、`net/http` 在 `response.go` 里可能变成未用（`WriteError` 用了 `net/http.StatusInternalServerError`，所以 `net/http` 还留着；`encoding/json` 会没人用）。`goimports`/编译器会报，删干净。
2. **成功出口误用 Abort**：`WriteSuccess` 用 `c.JSON` 就行，别用 `AbortWithStatusJSON`——虽然不出错，但语义上成功路径无需中断链。
3. **收敛后忘了删旧 helper**：`handler.writeError` 和 `writeUnauthorized` 删掉后要确认没有遗留调用点（编译器会兜底）。
4. **测试断言别再手拼信封字符串**：`response_test.go` 里断言 body 时，反序列化成 `ErrorBody`/`Body` 再比字段，别用 `strings.Contains` 拼——那正是 B4 要消灭的味道。
5. **改了响应写法但改变了行为**：本次是纯重构，端到端响应必须逐条对得上 Day 25——尤其 DELETE 成功仍是 `{"message":"ok"}`（`omitempty`）、401/500 文案不变。


