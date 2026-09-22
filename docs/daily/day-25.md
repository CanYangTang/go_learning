# Day 25 学习记录

## 日期

2026-09-20

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/25

## 今日教案

- 教案文档：`docs/daily/day-25-lesson.md`

## 核心任务

- 新包 `internal/auth`：用 `golang-jwt/jwt/v5`（HS256）签发/校验带过期时间的 JWT。
- `internal/middleware` 的 `RequireAuth`：解析 `Authorization: Bearer <token>`，缺失/非法/过期一律 `401`，走统一错误信封。
- 拆路由组：TODO 接口挪到受保护子组（backlog A6），`/health`、`register`、`login` 保持公开。
- `login` 成功返回 `token`。
- CORS 从「反射任意 Origin」收紧成白名单（backlog A5，加认证前的前置）。

## 范围决策

- 只做鉴权。**不做**按用户隔离 TODO 数据（`Todo.UserID`），所有登录用户仍共享同一份 todos。`UserID` 留到后续。

## 可选挑战题

- token 过期处理（核心已内建，重点是实测过期被拒）。
- 刷新 token（讨论为主，今天不实现）。

## 答疑记录

- 学习教案期间无额外提问，直接进入实现。

## Quiz

1. **删掉 keyfunc 里的 HMAC 方法断言，攻击者能怎么伪造能过校验的 token？为什么能绕过签名？**
   - 回答：攻击者可以把 alg 改成 "none"，递一个无签名的 token。
   - 结果：✅ 对（要点命中）。
   - 标准答案：主路是 alg=none（签名为空，不校验方法时库不验签，任何人自造 payload 都通过）；另一条是 RS256→HS256 算法混淆攻击（服务端 RSA 公钥公开，攻击者拿公钥当 HMAC 密钥签 HS256 token）。断言 `t.Method.(*jwt.SigningMethodHMAC)` 把这两条一起堵死。

2. **JWT payload 只是 base64、不是加密。说一条绝不能放、一条可以放的东西。**
   - 回答：绝对不能放密码，email 可以。
   - 结果：✅ 对。
   - 标准答案：判据是「公开出去也无所谓的身份标识」——可放 userID、过期时间、email；不能放密码、密钥、其他用户隐私。

3. **为什么把 m.Parse 的原始错误只写日志、只给客户端回笼统 "invalid or expired token"？直接回原始错误会多暴露什么？**
   - 回答：为了安全，暴露具体的错误原因。
   - 结果：🔶 部分对（方向对，没说清暴露了什么）。
   - 标准答案：原始错误会区分过期/签名错/格式错，这是给攻击者的情报（如 "expired" 说明结构和密钥本对、只是过期，缩小爆破面），还可能泄露内部实现；统一笼统回复掐掉这条旁路信息。

4. **protected := v1.Group("") 和 v1 的 URL 有何区别？为什么能保护 /todos 却不动 /health、/login？**
   - 回答：没有区别，可以再加一个 RequireAuth 的中间件。
   - 结果：🔶 部分对（URL 无区别✓，但没点出中间件只挂子组）。
   - 标准答案：两组 URL 前缀都是 /api/v1，无区别；关键在中间件只挂在子组上——/health、/login 注册在 v1 上不经过 RequireAuth，/todos 注册在 protected 子组上才有。Group("") = 同路径、不同中间件链。

5. **bearerToken 用 len(header) <= len(prefix)（而非 <）。举一个 <= 挡掉、< 漏过的例子，说明漏过后果。**
   - 回答：漏掉会引发安全问题。
   - 结果：🔶 不足（没给例子）。
   - 标准答案：prefix="Bearer " 长 7，当 header 正好 "Bearer "（只有前缀、空 token）时，<= 是 7<=7 为真直接挡掉；< 是 7<7 为假会放过，取 header[7:]="" 当合法 token 交给 Parse。危害有限（Parse("") 仍返 401），属 defense-in-depth——不该把空 token 传到下游。

- 得分：3/5，3~5 题差在「说清机制/给例子」。

## 今日产出

- 新包 `internal/auth`：`Manager.Generate/Parse`（HS256，`RegisteredClaims.Subject` 存 userID，含过期校验与 alg=none 防御）。
- `pkg/apperror`：新增 `Unauthorized`（UNAUTHORIZED/401）。
- `internal/middleware/auth.go`：`RequireAuth` + `bearerToken` + `writeUnauthorized` + `UserIDFromContext`；原始错误只入日志（带 request_id），客户端只收笼统 401。
- `internal/middleware/cors.go`：CORS 从反射任意 Origin 收紧为白名单（A5）。
- `internal/handler/user.go`：`Login` 成功签发 JWT，返回 `LoginResponse{id,email,token}`。
- `internal/config/env.go`：`JWTSecret()`（`JWT_SECRET` env + dev 兜底）、`AllowedOrigins()`（`CORS_ALLOWED_ORIGINS` 逗号分隔 + 兜底）。
- `cmd/server/main.go`：注入 `auth.Manager`；拆路由组——public: health/register/login，protected(`v1.Group("")`+RequireAuth): 5 条 todos（A6）。
- 测试：`auth`、`middleware`、`handler`、`apperror` 相关单测全部补齐并通过。
- `go.mod`：`golang-jwt/jwt/v5 v5.3.1` 提升为直接依赖。

## 运行过的命令

```bash
gofmt -l . && gofmt -w .
go mod tidy
go build ./...
go vet ./...
go test -count=1 ./internal/... ./pkg/...
RUN_INTEGRATION_TESTS=true go test -count=1 ./internal/repository/
# 端到端：register → login 取 token → /todos 带/不带/乱码 token → CORS 白名单 → OPTIONS 204
JWT_SECRET=e2e-test-secret ./server & curl ...
```

## 遇到的问题

- 用户填完后 auth.go 的 import 用了空格缩进，`gofmt -l` 命中，`gofmt -w` 修复并重排分组。
- 本次涉及函数上方的 `// TODO: implement` 脚手架注释在实现后清理（backlog C6 复现）；早几天遗留的同类注释（todo.go/repository/service）暂未清，待单独扫。
- `go mod tidy` 时机：golang-jwt 在真正被 import 前是 `// indirect`，必须在实现之后 tidy 才会提升为直接依赖。

## 关键收获

1. JWT 是「传输层身份」而非业务规则，独立成 `internal/auth` 包（不进 service），keyfunc 的 HMAC 方法断言是 alg=none / 算法混淆的强制防线。
2. 认证失败的原始错误是攻击者情报，只入日志（带 request_id）、客户端只收笼统 401；这与「重复键/账号枚举不泄露」是同一条安全原则。
3. `Group("")` 空路径子组 = 同 URL、不同中间件链，是「按路由分级鉴权」的干净做法，不改任何 URL。

## 明日计划

- Day 26：配置迁到环境变量（DB 密码、`.env.example`），连接池参数生效（backlog A4/B1/D5/E2）。
