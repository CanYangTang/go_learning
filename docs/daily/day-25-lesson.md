# Day 25 教案：JWT 鉴权 + 保护 TODO 路由

## 学习目标

学完今天，需要能够做到：

1. 说清 JWT 是什么、为什么用**对称签名（HS256）**、token 里放什么、不放什么。
2. 用 `github.com/golang-jwt/jwt/v5` 签发一个带过期时间的 token，并在中间件里校验它。
3. 实现 `internal/middleware` 的 `RequireAuth`：解析 `Authorization: Bearer <token>`，缺失/非法/过期一律返回 `401`，走统一错误信封。
4. 把 TODO 路由挪到**受保护的路由组**后面，`/health`、`/users/register`、`/users/login` 保持公开（解决 backlog A6）。
5. `login` 成功后在响应里返回 `token`。
6. 把 CORS 从「反射任意 Origin」收紧成**白名单**（解决 backlog A5）——这是加认证之前必须先做的事。

今天**不做**按用户隔离 TODO 数据（`Todo.UserID`）。范围已确认为「只做鉴权」：所有登录用户看到的仍是同一份 todos。`UserID` 留到后续某天。

---

## Day 25 的位置

Week 4 第四天。Day 23 加了 `user`（register/login，但 login 不发 token），Day 24 补齐了 TODO 的 CRUD。今天把这两条线接起来：**login 发 token，TODO 接口要带 token 才能访问**。

这天的架构意义：第一次出现「跨请求的身份」。在此之前每个请求都是无状态、匿名的；今天之后，一个请求会先经过 `RequireAuth` 证明「我是谁」，才能进入业务 handler。

---

## 第零件事：先收紧 CORS（backlog A5，必须在加认证之前）

现在的 `internal/middleware/cors.go` 把客户端传来的 `Origin` **原样回写**到 `Access-Control-Allow-Origin`：

```go
origin := c.GetHeader("Origin")
if origin == "" {
    origin = "*"
}
c.Header("Access-Control-Allow-Origin", origin)
```

在所有接口都公开时，这顶多是「谁都能跨域调」。但今天加了认证之后，语义变了：如果浏览器里已经登录的用户访问了恶意站点 `evil.com`，反射式 CORS 等于允许 `evil.com` 的脚本带着用户的 token 跨域读取受保护接口的响应。**必须先改成白名单。**

改法：`CORS` 接收一个允许的 Origin 列表，只有命中白名单的 `Origin` 才回写 `Access-Control-Allow-Origin`；不命中就不设这个头（浏览器随即拦截）。顺带补一个 `Access-Control-Max-Age`，让预检结果被浏览器缓存一段时间，减少 `OPTIONS` 次数。

```go
// CORS 现在接收白名单。命中才回写 ACAO；不命中不设，浏览器自然拦下。
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = true
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && allowed[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")
		c.Header("Access-Control-Max-Age", "600")
		c.Header("Vary", "Origin")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
```

几个决策点：

- **不设 `Access-Control-Allow-Credentials`。** 我们用 `Authorization: Bearer` 头传 token，不用 cookie。带凭证的 CORS（`Allow-Credentials: true`）另有一整套坑（不能配合 `*`、预检更严），我们不需要，就不开。
- **白名单从哪来？** 今天从 `main.go` 传入，默认 `http://localhost:3000`（一个假想的前端），用环境变量 `CORS_ALLOWED_ORIGINS`（逗号分隔）覆盖。这条链路和 `DB_DSN` 是同一个模式：默认值写死方便本地，环境变量兜生产。
- 不命中白名单时，其余 `Access-Control-*` 头设不设都无所谓——真正决定浏览器放不放行的是 `Allow-Origin` 那一个。设了也不构成泄漏，留着让代码简单。

> 这里改了 `CORS` 的签名，所以 `cmd/server/main.go` 的调用点和 `internal/middleware/middleware_test.go` 里的 CORS 测试都要跟着改。

---

## 第一件事：JWT 是什么，token 里放什么

JWT（JSON Web Token）是一段 `header.payload.signature` 三段式、用 `.` 连接的字符串。前两段是 base64url 编码的 JSON（**任何人都能解开读到**，不是加密！），第三段是签名。

- **header**：`{"alg":"HS256","typ":"JWT"}`——用什么算法签的。
- **payload**：claims，就是我们放进去的数据。用 `RegisteredClaims` 里的标准字段：`sub`（subject，放 user id）、`exp`（过期时间）、`iat`（签发时间）。
- **signature**：用密钥对 `header.payload` 做 HMAC-SHA256。**只有持有密钥的人能算出正确签名**，所以客户端改了 payload（比如把 `sub` 改成别人的 id），签名就对不上，校验失败。

关键认知：

1. **payload 不加密，只防篡改。** 别往 token 里放密码、密码哈希这类敏感信息——谁都能 base64 解开看。放个 user id、过期时间就够了。
2. **对称签名（HS256）**：签发和校验用**同一个密钥**。适合「同一个服务既发又验」的场景，正是我们的情况。（非对称的 RS256 用私钥签、公钥验，适合「A 发、B 验」的多服务场景，今天不需要。）
3. **密钥泄漏 = 任何人都能伪造任意用户的 token。** 所以密钥必须从环境变量来，不能写死进仓库。今天用 `JWT_SECRET` 环境变量，配一个开发用的默认值（和 DB 密码同样是「本地方便、生产必须覆盖」）。

---

## 第二件事：签发和校验 token（新包 `internal/auth`）

token 的签发（Generate）和校验（Parse）是一对，都依赖 JWT 库和密钥。把它们放进一个新包 `internal/auth`，理由和 Day 24「`:id` 解析放 handler」同源：**JWT 是一种传输层的身份机制，不是业务规则。** `service` 层不应该知道 token 长什么样——它只关心 email/password 对不对、user 存不存在。所以 token 的生死不放 service，单独成包。

这个包也不 import gin：它只跟字符串和 user id 打交道，谁调用它、怎么把 401 写回响应，是中间件的事。

`auth.Manager` 持有密钥和有效期，在 `main.go` 里构造一次，注入给「需要发 token 的 handler」和「需要验 token 的中间件」——依赖注入仍然只在 `main.go` 发生。

```go
package auth

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Manager 签发和校验 JWT。密钥和有效期在 main.go 注入。
type Manager struct {
	secret []byte
	ttl    time.Duration
}

func NewManager(secret []byte, ttl time.Duration) *Manager {
	return &Manager{secret: secret, ttl: ttl}
}

// Generate 为 userID 签发一个带过期时间的 token。
func (m *Manager) Generate(userID uint) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatUint(uint64(userID), 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// Parse 校验签名和过期时间，返回 token 里的 userID。
// 任何校验失败（签名错、过期、格式错、算法不对）都返回 error，
// 由中间件统一翻译成 401——这个包不碰 HTTP 语义。
func (m *Manager) Parse(tokenString string) (uint, error) {
	var claims jwt.RegisteredClaims
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		// 只接受 HMAC。不写这一步，攻击者可以把 alg 改成 "none"
		// 递一个无签名的 token 进来——这是 JWT 最经典的漏洞。
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return m.secret, nil
	})
	if err != nil {
		return 0, err
	}
	id, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
```

### 实测确认（临时程序，跑完即删）

上面这套写法的边界行为我逐条验证过：

```
[sign]      err=<nil> tokenLen=143
[valid]     err=<nil> subject="42"          ← 正常 token 解回 sub=42
[expired]   err="token has invalid claims: token is expired"
            errors.Is(err, jwt.ErrTokenExpired)=true   ← 过期自动被 ParseWithClaims 挡下
[badsig]    err="token signature is invalid: signature is invalid"
            errors.Is(err, jwt.ErrTokenSignatureInvalid)=true   ← 换个密钥就验不过
[malformed] errors.Is(err, jwt.ErrTokenMalformed)=true          ← 乱串直接 malformed
[none-parse] err="token is unverifiable: ... unexpected signing method: none"
            ← alg=none 被 keyfunc 里的 HMAC 检查挡住
```

两个要点：

1. **过期是 `ParseWithClaims` 自动校验的**，只要 claims 里有 `exp`，不用自己比时间。这就是「可选挑战题：token 过期处理」其实已经内建了——我们把 `exp` 放进去，过期就自动拒。
2. **`alg=none` 攻击**靠 keyfunc 里那句 `t.Method.(*jwt.SigningMethodHMAC)` 类型断言挡住。这一步不能省。

对中间件来说，这些错误**不需要区分**：签名错、过期、格式错，对客户端一律是「你这 token 不好使」= `401`。原始错误只写日志（按 request_id），不回给客户端——和 Day 21 的绑定错误、Day 23 的登录错误同一个原则：不给攻击者额外信息。

---

## 第三件事：`apperror.Unauthorized`

`pkg/apperror` 目前有 `Internal`/`Validation`/`NotFound`，还差 401。加一个：

```go
// Unauthorized 用于缺失或非法的凭证。
func Unauthorized(message string) Error {
	return New("UNAUTHORIZED", message, http.StatusUnauthorized)
}
```

`docs/api/todo-api.md` 的错误码表里 `UNAUTHORIZED / 401` 一直标着「计划中（Day 25）」，今天落地后改成「已实现」。顺带给 `pkg/apperror/error_test.go` 补一条 code/status 断言（Day 21 已经给其他构造函数补过测试，照抄一条即可）。

---

## 第四件事：`RequireAuth` 中间件

替换掉 `internal/middleware/auth.go` 里的 `AuthPlaceholder`（那个全局挂着、什么都不做的占位）。新的 `RequireAuth` 接收 `*auth.Manager`：

```go
package middleware

import (
	"log"
	"strings"

	"github.com/CanYangTang/go_learning/internal/auth"
	"github.com/CanYangTang/go_learning/pkg/apperror"
	"github.com/CanYangTang/go_learning/pkg/response"
	"github.com/gin-gonic/gin"
)

const userIDContextKey = "user_id"

// RequireAuth 校验 Authorization: Bearer <token>。缺失/非法/过期一律 401。
func RequireAuth(m *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token, ok := bearerToken(header)
		if !ok {
			writeUnauthorized(c, "missing or malformed authorization header")
			return
		}

		userID, err := m.Parse(token)
		if err != nil {
			// 原始错误（过期/签名错/格式错）只写日志，不回给客户端。
			log.Printf("request_id=%s auth_error=%v", RequestIDFromContext(c), err)
			writeUnauthorized(c, "invalid or expired token")
			return
		}

		// 存进 context，未来按用户隔离时 handler 直接取（今天还没用到）。
		c.Set(userIDContextKey, userID)
		c.Next()
	}
}

// bearerToken 从 "Bearer xxx" 里抠出 token，形状不对返回 false。
func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	return strings.TrimSpace(header[len(prefix):]), true
}

// writeUnauthorized 直接写 401 信封。handler.writeError 是未导出的、在别的包，
// 所以这里和 Recovery 一样，用 pkg/response 重建信封——形状共享，helper 不共享。
func writeUnauthorized(c *gin.Context, message string) {
	appErr := apperror.Unauthorized(message)
	c.AbortWithStatusJSON(appErr.StatusCode, response.ErrorBody{
		Error: response.ErrorPayload{Code: appErr.Code, Message: appErr.Message},
	})
}

// UserIDFromContext 供后续按用户隔离时使用。
func UserIDFromContext(c *gin.Context) (uint, bool) {
	v, exists := c.Get(userIDContextKey)
	if !exists {
		return 0, false
	}
	id, ok := v.(uint)
	return id, ok
}
```

要点：

- **为什么中间件自己写信封、不调 `writeError`？** `writeError` 是 `handler` 包里未导出的函数，而 `handler` import 了 `middleware`——反过来 import 会成环。`Recovery` 中间件早就踩过这条线，用 `pkg/response` 重建信封解决。照它的先例。
- **`c.Set(userIDContextKey, userID)`** 今天不会被读，但它是「身份沿请求往下传」的约定入口。留 `UserIDFromContext` 是给未来按用户隔离铺路，不是死代码——但如果你觉得现在加没用到的 helper 别扭，也可以今天先不加，等真要隔离时再加。这个取舍你定，定了在注释里说一句。
- **`bearerToken` 的大小写**：HTTP header value 里 `Bearer` 理论上大小写不敏感，用 `strings.EqualFold` 比前缀更稳。

---

## 第五件事：拆路由组——公开的 vs 受保护的（backlog A6）

现在 `AuthPlaceholder` 挂在 `router.Use`（全局）。一旦换成真的 `RequireAuth`，会**把 `/health` 也拦掉**——探针、`register`、`login` 都需要在没 token 时能访问。所以不能全局挂，要挂在**受保护的子组**上。

```go
// main.go 里，先构造 token manager
secret := []byte(config.JWTSecret())
tokenManager := auth.NewManager(secret, 24*time.Hour)

userHandler := handler.NewUserHandler(userService, tokenManager) // Login 要发 token

router := gin.New()
// 注意：AuthPlaceholder 从这里去掉了；鉴权改挂在受保护组上。
router.Use(middleware.RequestID(), middleware.Logging(), middleware.Recovery(), middleware.CORS(config.AllowedOrigins()))

v1 := router.Group("/api/v1")
{
	// —— 公开：不需要 token ——
	v1.GET("/health", handler.HealthHandler)
	v1.POST("/users/register", userHandler.Register)
	v1.POST("/users/login", userHandler.Login)

	// —— 受保护：必须带合法 token ——
	protected := v1.Group("")
	protected.Use(middleware.RequireAuth(tokenManager))
	{
		protected.POST("/todos", todoHandler.CreateTodo)
		protected.GET("/todos", todoHandler.ListTodos)
		protected.GET("/todos/:id", todoHandler.GetTodo)
		protected.PUT("/todos/:id", todoHandler.UpdateTodo)
		protected.DELETE("/todos/:id", todoHandler.DeleteTodo)
	}
}
```

`v1.Group("")` 建一个路径前缀不变、只是多挂一层中间件的子组——TODO 的 URL 还是 `/api/v1/todos`，没变，只是多了道 `RequireAuth` 门。这是 Gin 分组中间件的标准用法。

> `config.JWTSecret()` 和 `config.AllowedOrigins()` 是两个新的小 helper：前者读 `JWT_SECRET`（默认一个开发用串），后者读 `CORS_ALLOWED_ORIGINS` 逗号分隔（默认 `http://localhost:3000`）。和 `DefaultDSN()`/`DB_DSN` 一个模式。**注意日志里不要打印密钥本身。**

---

## 第六件事：`login` 返回 token

`UserHandler` 现在只持有 `service`。要发 token，得把 `tokenManager` 也注入进去。Login 里：先 `service.Login` 验密码（这步不变），拿到 `user` 后 `tokenManager.Generate(user.ID)` 签一个 token，塞进响应。

```go
type UserHandler struct {
	service UserService
	tokens  *auth.Manager // 新增
}

func NewUserHandler(service UserService, tokens *auth.Manager) *UserHandler {
	return &UserHandler{service: service, tokens: tokens}
}

// LoginResponse 比 UserResponse 多一个 token 字段。
type LoginResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Token string `json:"token"`
}

func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("request_id=%s bind_error=%v", middleware.RequestIDFromContext(c), err)
		writeError(c, apperror.Validation("invalid request body"))
		return
	}

	user, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		writeError(c, err)
		return
	}

	token, err := h.tokens.Generate(user.ID)
	if err != nil {
		log.Printf("request_id=%s token_error=%v", middleware.RequestIDFromContext(c), err)
		writeError(c, apperror.Internal("login failed"))
		return
	}

	c.JSON(http.StatusOK, response.Body{
		Data:    LoginResponse{ID: user.ID, Email: user.Email, Token: token},
		Message: "ok",
	})
}
```

`Register` 不发 token（注册完还得再 login，或者你也可以让注册直接发 token——今天不做，保持 register 只建账号）。

`handler` 现在要 import `internal/auth`。注意别让 `service` 也去 import `auth`——发 token 的动作留在 handler，service 只管「密码对不对」。

---

## 今日代码结构

新文件：

```
internal/auth/token.go          # Manager：Generate / Parse（新包）
internal/auth/token_test.go     # 单元测试：签发→解回、过期拒绝、篡改/换密钥拒绝、alg=none 拒绝
```

改动：

```
pkg/apperror/error.go           # 加 Unauthorized（UNAUTHORIZED / 401）
pkg/apperror/error_test.go      # 加一条 Unauthorized 的 code/status 断言
internal/middleware/cors.go     # CORS 改成收白名单 []string
internal/middleware/auth.go     # 删 AuthPlaceholder，加 RequireAuth + bearerToken + UserIDFromContext
internal/middleware/middleware_test.go  # 改 CORS 测试签名；删 AuthPlaceholder 测试，加 RequireAuth 测试
internal/handler/user.go        # UserHandler 加 tokenManager；Login 发 token；加 LoginResponse
internal/handler/user_test.go   # Login 测试断言响应里有 token
internal/config/env.go          # 新增 JWTSecret() / AllowedOrigins()（或加到现有 config 文件里）
cmd/server/main.go              # 构造 auth.Manager；拆公开/受保护路由组；CORS 传白名单；去掉 AuthPlaceholder
docs/api/todo-api.md            # 受保护接口标注需要 Bearer；login 响应加 token；错误码表 UNAUTHORIZED 转「已实现」；CORS 说明改成白名单
```

按惯例，AI 负责：包/结构体/接口声明、`main.go` 装配和路由、测试文件、带 `// TODO: implement` 和 `panic("not implemented")` 的空骨架。**你负责填**：`Generate`/`Parse` 的方法体、`RequireAuth` 的逻辑、`CORS` 白名单判断、`Login` 里签 token 的那几行。

---

## 今日验收标准

1. 不带 `Authorization` 头请求 `GET /api/v1/todos` → `401`，信封是 `{"error":{"code":"UNAUTHORIZED","message":"missing or malformed authorization header"}}`。
2. 带一个乱写的/过期的/被篡改的 token → `401`（`invalid or expired token`）。
3. `POST /api/v1/users/login` 成功 → 响应 `data` 里有非空 `token`。
4. 用上一步的 token 带 `Authorization: Bearer <token>` 请求 `GET /api/v1/todos` → `200`，正常返回列表。
5. `GET /api/v1/health`、`register`、`login` **不带 token 也能访问**（A6 不回归）。
6. CORS：白名单内的 `Origin` 回 `Access-Control-Allow-Origin`；白名单外的 `Origin` **不回**这个头（A5 修复）。
7. 三层边界不破：`internal/service` 不 import `auth`/`gin`；`internal/auth` 不 import `gin`/`gorm`。
8. `make fmt test vet` 全绿；`go mod tidy` 后 `golang-jwt/jwt/v5` 是直接依赖（不再是 `// indirect`）。
9. 端到端：`login` 拿 token → 带 token 走通一次完整 CRUD；不带 token 被 401 挡下。

---

## 可选挑战题

1. **token 过期处理**（Issue #25 原文的可选项）：其实核心已内建（`exp` + `ParseWithClaims` 自动校验）。挑战是**实测**——签一个 `ttl` 很短（比如 `2*time.Second`）的 token，等它过期再请求，确认返回 401 且日志里是 `token is expired`。可以临时把 `NewManager` 的 ttl 调小验证，验完调回。
2. **刷新 token**：加一个 `POST /api/v1/users/refresh`，用旧 token（未过期）换一个新 token。想清楚：刷新和登录的区别是什么？无限刷新会不会让「过期」形同虚设？（提示：真实系统用 access token 短 + refresh token 长的双 token，今天不做，只讨论。）

---

## 今天最容易踩的坑

1. **`AuthPlaceholder` 直接换成全局 `RequireAuth`**，把 `/health`、`login` 一起拦了——登录接口自己需要 401 才能拿 token，成了死锁。必须挂在**受保护子组**，公开接口留在外面。
2. **keyfunc 里不检查签名算法**，`alg=none` 攻击就能绕过。那句 `t.Method.(*jwt.SigningMethodHMAC)` 断言不能省。
3. **把 401 的原始错误（过期/签名错）回给客户端。** 只写日志，对外统一 `invalid or expired token`。
4. **密钥写死进代码或打进日志。** 从 `JWT_SECRET` 读，默认值仅供本地；任何日志都不要打印密钥。
5. **CORS 收窄后忘了改测试和 `main.go` 调用点**，编译不过——`CORS()` 变成了 `CORS([]string{...})`。
6. **让 `service` 去 import `auth` 发 token。** 发 token 是 handler 的活，service 只管验密码——别破坏分层。
7. **`go mod tidy` 时机**：`golang-jwt` 现在是 `// indirect`（还没有代码 import 它）。等 `internal/auth` 写完 import 之后再 tidy，它才会转成直接依赖；现在 tidy 会把它删掉。





