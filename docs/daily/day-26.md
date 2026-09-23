# Day 26 学习记录

## 日期

2026-09-22

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/26

## 今日教案

- 教案文档：`docs/daily/day-26-lesson.md`

## 核心任务

- 删死代码 B2：`pkg/response` 的 `JSON`/`Error`/`GinJSON`/`GinError`（后两个是 `c any` + 断言 + 静默返回的反模式）。
- 统一错误出口：`pkg/response.WriteError(c, err)` 作为唯一出口，收敛 `handler.writeError` + 两处 middleware 的重复信封构造。
- 对称成功出口：`pkg/response.WriteSuccess(c, status, data)`。
- 补 `pkg/response` 测试（C4）；`health_test.go` 换 `strings.Contains`（B4）。
- 验证并用编译期断言钉死分层边界（service/handler/auth 已实测干净）。

## 范围决策

- **只做重构**。配置迁环境变量（DB 密码、`.env.example`、连接池 A4/B1/D5/E2）留独立一天，不塞进今天。
- 错误出口设计：`pkg/response` 依赖 gin + apperror，提供 `WriteError(c *gin.Context, err error)`（已确认 import 图不成环）。

## 可选挑战题

- 配置加载改进（Issue 里标 optional）——本次不做，留后续。

## 答疑记录

- 本日无独立答疑（教案已覆盖 no-cycle 证明与 AbortWithStatusJSON 双角色验证），学完直接进实现。

## Quiz

满分 5/5。

**Q1｜`errors.As` vs 类型断言,遇到 `fmt.Errorf("%w", apperror.Validation(...))` 有何区别?**
- 回答:`errors.As` 会解包 `%w` 的错误链,类型断言只看最外层;包裹的 apperror 只有 `errors.As` 能识别,否则误判成 500。
- 结果:✓ 正确。
- 标准答案:`errors.As` 沿错误链逐层 `Unwrap` 匹配目标类型;`err.(apperror.Error)` 只断言最外层。包裹后外层是 `*fmt.wrapError`,断言失败 → 走 500 兜底,信息丢失。`TestWriteErrorWrappedAppErrorIsUnwrapped` 守此行为。

**Q2｜为何错误路径 `AbortWithStatusJSON`、成功路径 `c.JSON`?Abort 停掉了什么?**
- 回答:错误路径 Abort 是为了让共享的 `WriteError` 在 middleware 里阻止后续 handler 执行,成功路径在链尾无需中断。
- 结果:✓ 正确。
- 标准答案:Abort 把 `c.index` 推到 abortIndex,使后续 `c.Next()` 不再进入下游 handler —— middleware 角色(如 RequireAuth 写 401)必须掐断链条。handler 角色处于链尾,Abort 无害。成功响应只在链尾发生,没有下游要停。

**Q3｜非 apperror 错误(如 `connection refused: 10.0.0.5:3306`)客户端看到什么?为何不能返回原始文案?**
- 回答:对外只返回固定 `internal server error`;原始错误会泄露内网 IP、表名、路径等情报。
- 结果:✓ 正确。
- 标准答案:兜底返回 `{"error":{"code":"INTERNAL_ERROR","message":"internal server error"}}` + 500。原始 err 常含内网 IP / 表名 / SQL / 文件路径,是攻击者的侦察情报,只能进日志(按 request_id 关联),绝不出网。`TestWriteErrorNonAppErrorFallsBackTo500` 断言 `10.0.0.5` 不泄露。

**Q4｜`pkg/response` 引入 gin + apperror 为何不成环?**
- 回答:依赖方向单向 `handler/middleware → response → apperror → 标准库`,无环。
- 结果:✓ 正确。
- 标准答案:依赖单向:上游 handler/middleware 依赖 response,response 依赖 apperror(叶子,仅 import `net/http`),apperror 不回头依赖 response;gin 是第三方库、不 import 本项目。任何一环都不指回上游,故无循环。

**Q5｜`Data` 用 `json:"data,omitempty"`,DELETE `WriteSuccess(c, 200, nil)` 响应体?去掉 omitempty 呢?**
- 回答:omitempty 让 nil data 消失,返回 `{"message":"ok"}`;去掉则返回 `{"data":null,"message":"ok"}`。
- 结果:✓ 正确。
- 标准答案:与回答一致。`TestWriteSuccessNilDataIsOmitted` 断言响应体不含 `"data"`。

## 今日产出

- `pkg/response/response.go`:删死代码 B2(`JSON`/`Error`/`GinJSON`/`GinError`/`writeJSON`),新增唯一出口 `WriteError`(errors.As 拆 apperror,兜底 500 固定文案)+ `WriteSuccess`(c.JSON + `Body` 信封)。
- `pkg/response/response_test.go`(新):5 个测试覆盖 apperror 命中 / wrapped 解包 / 非 apperror 兜底且不泄露 / 成功包信封 / nil data 省略。
- 收敛调用点:删 `handler.writeError`、`middleware.writeUnauthorized`、`recovery` 手拼信封;handler 成功响应改走 `WriteSuccess`;`NotFoundHandler` 改走 `WriteError`。
- B4:`health_test.go` 改用 `strings.Contains`,删 `containsString`/`findSubstring`。
- 分层边界:`cmd/server/main.go` 加 4 条编译期 `var _ Iface = (*Concrete)(nil)` 断言。

## 运行过的命令

```bash
gofmt -l .                                   # clean
go build ./... && go vet ./...               # OK
go test -count=1 ./internal/... ./pkg/...    # 全绿（含 response 新增 5 测试）
RUN_INTEGRATION_TESTS=true go test -count=1 ./internal/... ./pkg/...  # 全绿
# 端到端 curl：health / 401 / register / login / create 201 / 400 / 404 / 坏 token 401
```

## 遇到的问题

- 填完后 4 空格缩进被 `gofmt -l` 标记 → `gofmt -w` 转 tab;并删掉两函数上方已过时的 `// TODO: implement` 注释块和 import 里的临时注释。

## 关键收获

1. `errors.As` 是错误链的正确匹配方式,类型断言遇 `%w` 包裹会漏判 —— 统一出口才能保证包裹后的 apperror 仍映射到正确状态码。
2. 一个 `AbortWithStatusJSON` 出口同时服务 handler 与 middleware 两种角色:Abort 在中间件掐断下游、在链尾无害。
3. 错误信息分层:结构化 apperror 出网,原始 err 只进日志 —— 内网 IP / 表名 / SQL 是侦察情报,固定 500 文案是安全边界。


## 明日计划

- Day 27 或后续：配置迁环境变量（DB 密码 → env、`.env.example`、连接池生效 A4/B1/D5/E2）；工程收尾 C3/C6（`make` 加 `-count=1`/`goimports`/lint、清早期遗留 TODO 注释）。
