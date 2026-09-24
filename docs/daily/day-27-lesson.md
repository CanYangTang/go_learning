# Day 27 教案：补测试覆盖缺口 + 刷新 API 文档

> Issue #27 *Add tests and API docs*。范围决策：**API 文档刷新 + 补齐纯单元测试缺口**（不碰需要真实 DB 的 repository/ConnectGorm，那些走集成测试已覆盖）。分工：**AI 全写测试和文档，你审查**。production 代码今天没有待填的空函数。

---

## 0. 今日全景

三周下来项目已经有 16 个测试文件、总覆盖率 76.9%。今天不是从零写测试，而是**收尾**：把几处能用纯单元测试补上的分支补满，再把 Day 21 写的 API 文档更新到 Day 25/26 的真实形状（加鉴权端点、换统一信封）。

两件事：
1. **测试**：补 `config` 的两个 env 函数、`middleware` 的两处分支和未导出函数、`handler` 的一处分支。
2. **文档**：`docs/api/todo-api.md` 重新对齐真实接口。

---

## 1. 覆盖率现状与缺口分类

`go test -cover` + `go tool cover -func` 是"哪里没测到"的唯一权威来源，不要凭感觉：

```bash
go test -count=1 -coverprofile=cov.out ./internal/... ./pkg/...
go tool cover -func=cov.out | awk '$3 != "100.0%"'   # 只看没满的
go tool cover -html=cov.out                          # 浏览器里红=未覆盖
```

当前 <100% 的函数分三类：

| 类别 | 函数 | 现状 | 今天怎么办 |
|---|---|---|---|
| **纯函数/env**（好测） | `config.JWTSecret` `config.AllowedOrigins` | 0% | 补单元测试（`t.Setenv`） |
| **分支没走全**（好测） | `middleware.Logging` 90.9%、`UserIDFromContext` 80%、`request_id` 的 `fallbackRequestID` 0%/`newRequestID` 75% | 缺一条分支 | 补 case 走到那条分支 |
| **需 DB / 防御性**（今天不碰） | `repository.*` 0%、`config.ConnectGorm`/`Connect`、`handler.Login` 80% 的 token 失败分支 | 集成测试覆盖 或 强行覆盖成本过高 | 见 §6 |

---

## 2. Go 测试基础回顾（本项目已在用的写法）

- **同包白盒测试**：本项目所有 `_test.go` 都是 `package middleware` / `package config`（不是 `xxx_test`），所以能直接调**未导出**函数（`fallbackRequestID`、`bearerToken`）。这就是为什么 §4 能直接测私有函数。
- **表驱动 + 子测试**：一组输入用 `[]struct{...}` 列出，`for _, tt := range cases { t.Run(tt.name, func(t *testing.T){...}) }`。子测试名会出现在失败输出里，定位快。
- **`httptest` + `gin.CreateTestContext`**：不起真 server，`rec := httptest.NewRecorder(); c, _ := gin.CreateTestContext(rec)`，直接喂 `c` 给中间件/handler，再断言 `rec.Code` / `rec.Body`。`pkg/response/response_test.go` 就是这个套路。

---

## 3. 测「读环境变量」的函数：`t.Setenv`

`config.JWTSecret()` 和 `config.AllowedOrigins()` 的行为**取决于环境变量**。测这种函数不能直接 `os.Setenv`（会污染后续测试），要用 `t.Setenv`：

```go
func TestJWTSecret(t *testing.T) {
	// 不设 -> dev 默认（注意：得先确保没有外部注入，用 os.Unsetenv 兜底）
	os.Unsetenv("JWT_SECRET")
	if got := JWTSecret(); got != devJWTSecret {
		t.Fatalf("默认应回退到 devJWTSecret, got %q", got)
	}
	// 设了 -> 用环境值
	t.Setenv("JWT_SECRET", "prod-key")
	if got := JWTSecret(); got != "prod-key" {
		t.Fatalf("应返回环境值, got %q", got)
	}
}
```

**为什么 `t.Setenv` 而不是 `os.Setenv`**：
- 测试结束自动恢复原值 —— 不会泄漏到别的测试。
- 一旦调用 `t.Setenv`，该测试**禁止 `t.Parallel()`**（改全局 env 与并行互斥，Go 会 panic 提示）—— 这是语言层面帮你挡掉的坑。

`AllowedOrigins` 的边界（已实测确认）：`" a.com , ,b.com "` → `["a.com","b.com"]`，即**逐段 `TrimSpace` 且丢弃空段**。测试要覆盖：空环境→默认单元素、逗号分隔多值、含空白和空段的清洗。

---

## 4. 白盒测未导出函数：`request_id.go`

```go
func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fallbackRequestID()   // 75%: 这条 err 分支正常跑不到
	}
	return hex.EncodeToString(b[:])  // 正常路径
}
func fallbackRequestID() string { return "request-id" }
```

同包测试能直接调：

```go
func TestNewRequestIDIsHex32(t *testing.T) {
	id := newRequestID()
	if len(id) != 32 { t.Fatalf("16 字节 hex 应为 32 字符, got %d", len(id)) }
	if _, err := hex.DecodeString(id); err != nil { t.Fatalf("应为合法 hex: %v", err) }
}
func TestFallbackRequestID(t *testing.T) {
	if fallbackRequestID() != "request-id" { t.Fatal("兜底值应为固定 request-id") }
}
```

`newRequestID` 的 `rand.Read` 失败分支属于**几乎不可能触发**（读系统 CSPRNG 失败意味着系统已崩），不为它做依赖注入 —— 见 §6。但 `fallbackRequestID` 本身可以直接测到，把它从 0% 拉满。

---

## 5. 把「没走全的分支」走全

**`Logging` 90.9%** —— 缺"没有 request_id"那条日志分支：

```go
if requestID != "" {
	log.Printf("... request_id=%s ...")   // 已测（RequestID 中间件在前，总有 id）
	return
}
log.Printf("...")                         // 没测到：单独跑 Logging、不挂 RequestID
```

补一个只挂 `Logging()`、不挂 `RequestID()` 的 case，`RequestIDFromContext` 返回空串 → 走到第二条 `log.Printf`。断言重点是**不改响应**（status/body 不变），日志内容不是断言对象。

**`UserIDFromContext` 80%** —— 三态要测全：

```go
func UserIDFromContext(c *gin.Context) (uint, bool) {
	v, exists := c.Get(userIDContextKey)
	if !exists { return 0, false }   // ① key 不存在
	id, ok := v.(uint)               // ③ 存了但类型不对
	return id, ok
}
```

- ① 不 `Set` 直接取 → `(0,false)`
- ② `c.Set(userIDContextKey, uint(7))` → `(7,true)`（happy path，可能已被 RequireAuth 测间接覆盖，但显式测更清晰）
- ③ `c.Set(userIDContextKey, "not-a-uint")` → 类型断言失败 → `(0,false)`

> `userIDContextKey` 是未导出常量，同包测试可直接引用。

---

## 6. 明知难覆盖的分支：不为 100% 而污染代码

两处 <100% **今天故意不追**：

1. **`handler.Login` 80%**：未覆盖的是 `h.tokens.Generate` 失败分支。`tokens` 是具体类型 `*auth.Manager`，HS256 签名实际上不会失败。要覆盖必须把 `tokens` 抽成 interface 再注入一个"必失败"的假实现 —— 这是**为了测试而改动 production 代码**，得不偿失。
2. **`newRequestID` 的 `rand.Read` 失败**：同理，要注入假随机源。

原则：**覆盖率是工具不是 KPI**。防御性分支（"这几乎不可能发生，但真发生了要有兜底"）留着兜底逻辑本身就是价值，为触发它而引入接口和 mock 会让代码为测试扭曲。YAGNI 同样适用于测试。审查时如果看到为了拉满覆盖率而把简单具体类型改成接口，要警惕。

---

## 7. API 文档刷新：`docs/api/todo-api.md`

Day 21 写的文档停在"无鉴权 + 旧响应形状"，Day 25 加了注册/登录/Bearer 鉴权、Day 26 统一了信封。今天对齐真实接口。文档要素：

- **基础信息**：base URL `/api/v1`、统一响应信封、`X-Request-ID` 响应头（每个响应都带）。
- **成功信封**：`{"data": <payload>, "message": "ok"}`；`data` 为 `null` 时被 `omitempty` 省略（如 DELETE → `{"message":"ok"}`）。
- **错误信封**：`{"error": {"code": "...", "message": "..."}}`，附错误码表（`VALIDATION_ERROR` 400 / `UNAUTHORIZED` 401 / `NOT_FOUND` 404 / `INTERNAL_ERROR` 500）。
- **公开端点**：`GET /health`、`POST /users/register`、`POST /users/login`。
- **受保护端点**（需 `Authorization: Bearer <token>`）：`POST/GET /todos`、`GET/PUT/DELETE /todos/:id`。
- **示例全部从运行中的服务实抓**（不是手编）—— 每个端点给一条可直接粘贴的 `curl`，含真实的请求体和响应体。这是本项目文档的铁律（D1 教训：文档写了 7 个不存在的接口）。

抓取方式：本地 `go run ./cmd/server`，按 §"运行命令"的 curl 序列跑一遍，把真实响应贴进文档。

---

## 8. 分工

| 谁 | 做什么 |
|---|---|
| **AI** | 写全部新测试（`config` env、`middleware` 分支+未导出函数、`handler` 可补分支）；重写 `docs/api/todo-api.md`；跑覆盖率验证。 |
| **你** | **审查**：测试是否真的走到目标分支（对着 `go tool cover -html` 看红转绿）、断言是否有意义（不是只断言"不 panic"）、文档示例是否与实抓响应一致、§6 的两处未覆盖是否认同不追。 |

（今天角色对调：平时你写实现我审查，今天我写你审查。）

---

## 9. 验收标准

1. `config.JWTSecret` / `config.AllowedOrigins` 覆盖率 100%（含默认回退 + 环境覆盖 + 空白清洗）。
2. `middleware.fallbackRequestID` 100%、`newRequestID` happy path 已测、`Logging` 两条日志分支都走到、`UserIDFromContext` 三态测全。
3. 新测试不依赖真实 DB、不依赖外部网络、可 `go test -count=1` 稳定重跑。
4. `docs/api/todo-api.md` 覆盖全部 8 个真实端点，含统一信封说明、错误码表、`X-Request-ID`、每端点可粘 `curl`，示例来自实抓。
5. `gofmt -l .` 干净、`go vet ./...` 干净、`go test -count=1 ./internal/... ./pkg/...` 全绿。
6. §6 两处防御分支保持现状，文档/注释说明为何不追。

---

## 10. 陷阱清单

1. **`os.Setenv` 泄漏**：在测试里用 `os.Setenv` 不恢复，会让后跑的测试读到脏值。一律用 `t.Setenv`（且该测试不能 `t.Parallel`）。
2. **断言无意义**：只断言"没 panic / status==200"而不查 body/分支，覆盖率涨了但没测到行为。每个 case 要有针对性断言。
3. **为覆盖率改 production 代码**：把具体类型改成 interface 只为注入失败 —— 见 §6，别做。
4. **文档手编示例**：凭记忆写请求/响应，容易和真实形状漂移（D1 就是这么烂的）。必须实抓。
5. **`omitempty` 的文档描述**：DELETE 成功是 `{"message":"ok"}` 而非 `{"data":null,...}`，文档别写错。
6. **白盒 vs 黑盒**：本项目测试与被测代码同包（能测私有函数）；若某天改成 `package xxx_test`，就只能测导出符号了 —— 今天维持同包。
