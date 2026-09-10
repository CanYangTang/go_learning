# Day 21 教案：Week 3 收尾与复盘

## 学习目标

学完今天，需要能够做到：

1. 说清 Week 3 到底建成了什么，以及还缺什么。
2. 真实跑通一次 TODO API 雏形，用 `curl` 确认数据落库。
3. 识别「文档漂移」，理解它为什么比没有文档更危险。
4. 修掉四处真实缺陷：缺失的 Recovery middleware、`NoRoute` 的 404 响应形状、绑定错误原文外泄、未 tidy 的 `go.mod`。
5. 把 `docs/api/todo-api.md` 从「愿望清单」改成「现状 + 计划」。
6. 给架构文档补上依赖方向和接口归属的说明。
7. 完成 `docs/weekly/week-03.md` 复盘。

---

## Day 21 的位置

Day 21 是 Week 3 的最后一天，主题是周总结，不是新知识。

Week 3 的五天分别落了这些地基：

- Day 15、16：Gin 路由组、请求绑定、统一响应信封。
- Day 17：MySQL 连接配置和 Docker Compose。
- Day 18：GORM 模型和 repository，含事务练习。
- Day 19：logging / CORS / request ID / auth 占位四个中间件。
- Day 20：handler → service → repository 分层，依赖在 main 统一组装。

Week 3 检查点里有一条是「TODO API 雏形可以运行」。这一条到现在**还没有被真正验证过** —— Day 20 收尾时本机 Docker 没启动，端到端只跑到单元测试为止。今天先把这个补上。

---

## 第一件事：真的跑一次

单元测试证明不了「数据真的落库了」。service 测试用的是内存假 repository，handler 测试用的是假 service，两者都没碰过 MySQL。

验证步骤：

```bash
colima start
docker compose -f deployments/docker-compose.yml up -d
make run
```

另开一个终端：

```bash
curl -sS -D - -X POST http://127.0.0.1:8080/api/v1/todos \
  -H 'Content-Type: application/json' \
  -d '{"title":"Week 3 done"}'

curl -sS -D - http://127.0.0.1:8080/api/v1/todos
```

要确认四件事：

1. POST 返回 `201`，响应里的 `id` 不是 0。
2. GET 能读回刚才那条，说明它真的进了 MySQL 而不是内存。
3. 重启一次服务再 GET，数据还在。
4. 两个响应都带 `X-Request-Id` 和 `Access-Control-*` 头。

第 3 条是关键。前两条在纯内存实现下也会通过，只有重启后数据还在才能证明持久化生效。

---

## 第二件事：文档漂移

`docs/api/todo-api.md` 是 Day 1 阶段按计划写的，用的是现在时语气，读起来像这些接口都已经存在了。实际对照代码，差得很远。

我把两份文档和当前代码逐条比对过，结果如下。

### 文档写了但代码没有

| 文档里的接口 | 现实 |
|--------------|------|
| `POST /api/v1/users/register` | 不存在，连 `model.User` 都还没有（Day 23） |
| `POST /api/v1/users/login` | 不存在（Day 23） |
| `GET /api/v1/todos/{id}` | 没有路由，`repository.FindByID` 写好了但没人调用（Day 24） |
| `PUT /api/v1/todos/{id}` | 没有路由，`repository.Update` 同上（Day 24） |
| `DELETE /api/v1/todos/{id}` | 没有路由，`repository.Delete` 同上（Day 24） |
| JWT Bearer 保护所有 TODO 接口 | `AuthPlaceholder` 是纯透传，`go.mod` 里连 JWT 库都没有（Day 25） |
| 分页参数 `page` / `page_size` / 筛选 `status` | 一个都没读取，`FindAll` 全表查（Day 24） |

### 字段和形状对不上

| 文档 | 代码 |
|------|------|
| `GET /health` | 实际是 `GET /api/v1/health`，裸 `/health` 会走到 404 |
| health 响应只有 status / service | 实际还有 `version: "0.1.0"` |
| 列表 `data` 是 `{items, page, page_size, total}` 对象 | 实际是**裸数组** |
| todo 对象有 `description`、`status`、`due_date` | 实际只有 `{id, title, done}`，状态是 **bool `done`** 不是字符串 `status` |
| 创建请求有 `title`、`description`、`due_date` | 只绑定 `title`，其余字段被静默丢弃 |
| 错误码规划了 5 个 | 只有 `VALIDATION_ERROR` 和 `INTERNAL_ERROR` 有构造函数并真的被返回过 |

### 代码做了但文档没提

- `X-Request-ID` 的透传/生成，以及响应头回写。
- CORS 的四个响应头、`Vary: Origin`、`OPTIONS` → 204。
- 访问日志格式 `request_id=… method=… path=… status=… latency=…`。
- `DB_DSN` 环境变量，以及默认连本地 Docker 库。
- 启动时 `AutoMigrate`。
- 软删除：`model.Todo` 带 `gorm.DeletedAt`，所以 `Delete` 不是真删，`FindAll` 会自动排除已删记录。
- Day 20 最重要的设计决策——接口定义在调用方、依赖在 main 统一组装——架构文档里一个字都没写。

### 为什么漂移比没文档更危险

没有文档时，别人会去读代码，读到的是事实。

有一份看起来完整、语气确定的过时文档时，别人会信它。前端照着 `status: "pending"` 写代码，实际拿到的是 `done: false`；照着 `data.items` 取列表，实际 `data` 就是数组；带上 `Authorization` 头以为受保护，其实谁都能调。这些错都发生在联调阶段，而不是读文档的时候。

修法不是删掉文档，而是把语气从「是」改成「现在是 / 计划是」。已实现的写现状，未实现的标上计划日期。

---

## 第三件事：今天的代码改动

审查文档的同时我把整个项目也过了一遍，得到一份 24 条的问题清单，落在 `docs/issues-backlog.md`，按计划 Day 分配了归属。今天动手修其中 4 条 —— 它们是同一个主题：**让所有异常出口收到同一个形状里**。

### 改动一：`NoRoute` 的响应形状（清单 A2）

`cmd/server/main.go` 里的兜底路由是这样：

```go
router.NoRoute(func(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"error": "not found",
	})
})
```

问题在响应形状。项目里所有其他错误都走 `handler.writeError`，出去的是：

```json
{"error": {"code": "VALIDATION_ERROR", "message": "title is required"}}
```

`error` 是一个对象。而 `NoRoute` 出去的是：

```json
{"error": "not found"}
```

`error` 是一个字符串。客户端只要写了统一的错误解析——比如读 `resp.error.code`——碰到任何未知路径就会解析失败或拿到空值。而未知路径恰恰是最容易被触发的错误：拼错 URL、忘了 `/api/v1` 前缀、直接访问文档里写的 `/health`。

### 怎么修

需要两步。

第一步，`pkg/apperror/error.go` 里补一个构造函数。现在只有 `Internal` 和 `Validation`：

```go
func NotFound(message string) Error {
	return New("NOT_FOUND", message, http.StatusNotFound)
}
```

`NOT_FOUND` 这个错误码在 `docs/api/todo-api.md` 的错误码表里已经规划了，只是一直没有实现。

第二步，把 404 的输出接到 handler 层统一的错误出口上。`writeError` 现在是 `internal/handler/todo.go` 里的包级私有函数，`NoRoute` 也在同一个包能用得上——新增一个 `internal/handler/notfound.go`：

```go
func NotFoundHandler(c *gin.Context) {
	writeError(c, apperror.NotFound("route not found"))
}
```

然后 `main.go` 变成一行：

```go
router.NoRoute(handler.NotFoundHandler)
```

这样 `net/http` 在 `main.go` 里就不再需要了，import 要跟着删。

### 一个值得注意的实测结论

`NoRoute` 的 handler **会经过全局 middleware**。我实测过：请求一个不存在的路径，响应里 `X-Request-Id` 和 `Access-Control-Allow-Origin` 都在，状态码 404。

这不是显然的——`NoRoute` 不属于任何路由组，很容易以为它绕过了 `router.Use`。Gin 的实现是把全局 middleware 链接在 `NoRoute` handler 前面，所以 404 也会被记日志、也带 request ID。这是好事：客户端报「404 了」时，你能靠 request ID 在日志里找到那一条。

### 改动二：补上 Recovery middleware（清单 A1，今天最严重的一条）

Day 19 挂 middleware 时我们用的是 `gin.New()`，然后自己挂了四个。`gin.New()` 和 `gin.Default()` 的区别就是后者自带 `gin.Logger()` 和 `gin.Recovery()` —— 我们用自己的 `Logging` 替掉了前者，**但 `Recovery` 从来没补回来**。

我实测了一下 handler 里 panic 会发生什么，结果比预想的糟：

```go
r.GET("/boom", func(c *gin.Context) { panic("boom") })
```

```text
RESULT: client error = Get "http://127.0.0.1:59086/boom": EOF
```

三个后果：

1. 客户端拿到的是 `EOF`，**连接被直接丢弃** —— 不是 500，更不是统一错误信封。前端只会看到「网络错误」，无从判断是自己的问题还是服务端的。
2. panic 从 `c.Next()` 里穿过 Logging middleware 一路往上抛，`c.Next()` 后面的代码根本没执行，所以**这次请求不会产生任何访问日志**。日志里查不到，等于这次请求没发生过。
3. 进程本身活着 —— `net/http` 对每个连接有一层 `defer recover()` 兜底，只在 stderr 打一行 `http: panic serving`。所以这个问题不会让服务挂掉，也就更不容易被发现。

第 3 条是它危险的原因：一个能让服务崩掉的 bug 你第一天就会修，一个只是偶尔丢连接、日志里毫无痕迹的 bug 可以潜伏很久。

写法上要注意两点，两点我都先写错了、实测后才改对，所以这里把实测结果一起放出来。

**第一点：`Recovery` 要挂在 `Logging` 之后（更内层）。**

我一开始的直觉是「Recovery 要在最外层才能兜住一切」，所以写成 `RequestID, Recovery, Logging`。实测发现这样虽然能返回 500，但**访问日志那一行还是没有**：

```text
Recovery 在外   status=500 日志 1 行
    | panic request_id=3cceb0a1... value=boom

Logging 在外    status=500 日志 2 行
    | panic request_id=fad86013... value=boom
    | request_id=fad86013... method=GET path=/boom status=500 latency=0s
```

原因是 middleware 的「外层」意味着更早进入、更晚返回。panic 从 handler 往外掀栈，谁的 `defer` 先接住，栈就在那里停住 —— 更外层的 middleware 里 `c.Next()` 之后的代码永远不会执行。所以 `Recovery` 在 `Logging` 外层时，`Logging` 的 post-`c.Next()` 代码已经被掀掉了，日志自然没有。反过来把 `Recovery` 放在 `Logging` 里面，`Recovery` 接住 panic 后正常返回，控制权回到 `Logging` 的 `c.Next()` 之后，那一行 `status=500` 就记上了。

正确顺序：

```go
router.Use(middleware.RequestID(), middleware.Logging(), middleware.Recovery(), middleware.CORS(), middleware.AuthPlaceholder())
```

`RequestID` 仍在最外面，这样 `Recovery` 记日志时能拿到 request ID；`CORS` 在 `Recovery` 里面，所以 panic 返回的 500 依然带跨域响应头。

**第二点：`Abort` 到底有没有用 —— 我这里错了两次，第二次是学习者纠正的。**

第一版我写「不 `Abort` 后续 handler 会继续跑」。实测了一种场景（业务 handler 里 panic），发现两种写法响应完全一样，于是改成第二版：「`Abort` 只是惯例，不是修 bug」。

第二版也是错的 —— 错在我只测了一种场景就推广了。正确答案取决于 **panic 发生在哪里**：

```text
panicBeforeNext  abort=true  status=500 后续中间件=false 路由handler=false body={"code":"INTERNAL_ERROR"}
panicBeforeNext  abort=false status=500 后续中间件=true  路由handler=true  body={"code":"INTERNAL_ERROR"}route ok
panicAfterNext   abort=true  status=200 后续中间件=true  路由handler=true  body=route ok{"code":"INTERNAL_ERROR"}
panicAfterNext   abort=false status=200 后续中间件=true  路由handler=true  body=route ok{"code":"INTERNAL_ERROR"}
```

原因在 `c.Next()` 的实现：它是 `c.index++` 然后循环 `c.handlers[c.index](c)`。panic 解栈后控制流回到**外层**的那个 `Next()` 循环，循环继续 `c.index++`。

- panic 发生在某个中间件**调用 `c.Next()` 之前**（此时 `c.index` 还停在那个中间件上）：恢复后 `index++` 落在下一个 handler 上，**后续中间件和业务 handler 会照常执行**。响应变成 `{"code":"INTERNAL_ERROR"}route ok` —— 两段 body 拼在一起，客户端拿到的是坏 JSON。`Abort` 把 `index` 设成 63，这才是它在阻止后续执行。**这种情况下 `Abort` 是必需的，不是惯例。**
- panic 发生在业务 handler（handler 链的最后一环）：`c.index` 已经推到末尾，恢复后 `index++` 越界，循环自然结束，加不加 `Abort` 都一样。我第一次只测了这一种。

顺带看第三、四行：panic 发生在某个中间件 `c.Next()` **之后**的收尾代码里时，业务 handler 已经写完 200 和响应体了，`Recovery` 再写只能追加 —— **状态码停在 200**，body 变成 `route ok{"code":"INTERNAL_ERROR"}`，gin 会打 `Headers were already written` 警告。`Recovery` 兜不住这种情况，`Abort` 也救不了。

结论：写 `AbortWithStatusJSON` 而不是 `c.JSON`。

方法论上的教训比结论更值钱：**测一种场景就下结论，等于把「我想到的那种情况」当成了全部。** 这条 panic 位置有三种（中间件 Next 前 / 业务 handler 内 / 中间件 Next 后），三种行为都不一样，我两次都只测了一种。

顺手可以复用同一个错误出口：`writeError(c, apperror.Internal("internal server error"))`。但注意 `writeError` 在 `internal/handler` 包，middleware 包调不到私有函数 —— 这里要么直接写 `response.ErrorBody`，要么把统一出口提到一个两边都能用的地方。这是个真实的设计取舍，实现时会撞上。

### 改动三：绑定错误别把原文抛出去（清单 A3）

`internal/handler/todo.go` 现在这样处理绑定失败：

```go
if err := c.ShouldBindJSON(&req); err != nil {
	writeError(c, apperror.Validation(err.Error()))
	return
}
```

`err.Error()` 直接进了对客户端的响应。三种坏输入的实测结果：

```text
{}            → "message":"Key: 'CreateTodoRequest.Title' Error:Field validation for 'Title' failed on the 'required' tag"
{"title":     → "message":"unexpected EOF"
[]            → "message":"json: cannot unmarshal array into Go value of type handler.CreateTodoRequest"
```

`CreateTodoRequest`、`handler.CreateTodoRequest` 是内部的 Go 结构体名，不该出现在 HTTP 响应里。这属于信息泄漏——不严重，但它免费告诉了攻击者你的内部结构。而且对正常客户端也没用：前端拿到 `unexpected EOF` 不知道该改什么。

修法是两句话：对外返回固定文案，原始错误只写日志。

```go
if err := c.ShouldBindJSON(&req); err != nil {
	log.Printf("request_id=%s bind failed: %v", middleware.RequestIDFromContext(c), err)
	writeError(c, apperror.Validation("invalid request body"))
	return
}
```

这里有个连带影响：`internal/handler/todo_test.go` 里两个断言只检查了 `"code":"VALIDATION_ERROR"`，没检查 message，所以改完测试**照样会过**。这正是这个问题一直没被发现的原因 —— 顺便把 message 的断言补上。

### 改动四：`go mod tidy`（清单 C1）

`go mod tidy -diff` 显示 `gin`、`gorm`、`gorm.io/driver/mysql`、`go-sql-driver/mysql` 四个直接依赖全被标成了 `// indirect`。这是之前 `go get` 之后没跑 tidy 留下的。

`// indirect` 的含义是「没有任何代码直接 import 它，只是别的依赖需要」。标错了不影响编译，但会误导任何想搞清楚项目直接依赖了什么的人 —— 包括几周后的你自己。一条命令解决：

```bash
go mod tidy
```


---

## 第四件事：文档怎么改

### `docs/api/todo-api.md`

改成两段结构：

- **已实现**：`GET /api/v1/health`、`POST /api/v1/todos`、`GET /api/v1/todos`、404 兜底。字段、状态码、响应形状全部按代码写，包括 `data` 是裸数组、todo 只有 `{id, title, done}`。
- **计划中**：其余接口保留，但每条标注计划日期（Day 23 用户与认证、Day 24 完整 CRUD 与分页、Day 25 JWT）。

再补一节「通用约定」，写现在真实存在但文档没提的东西：`X-Request-ID`、CORS 响应头、错误信封形状、当前真正会返回的错误码。

判断标准很简单：文档里的每个 `curl` 例子，现在跑一遍应该得到文档里写的结果。做不到的就得移到「计划中」。

### `docs/architecture/todo-api.md`

要修的过时描述：

- 「初期可使用内存实现」——现在是 GORM + MySQL，没有内存实现。
- User 模型——还不存在。
- TODO 表结构缺 `done` 和 `deleted_at`。

要补的是 Day 20 定下来但没记录的三条设计决策：

1. 依赖方向单向：handler → service → repository → model。反向依赖是禁止的，检验方式是 service 不能 import `gin` 或 `gorm`，handler 不能 import `gorm`。
2. 接口定义在调用方：`TodoRepository` 在 `internal/service`，`TodoService` 在 `internal/handler`，各自只声明用得到的方法。
3. 依赖只在 `cmd/server/main.go` 组装，三层内部都不自己 new 依赖。

这三条是整个项目可测性的来源，不写进文档，下次重构就会被破坏。

---

## 第五件事：`docs/weekly/week-03.md`

模板有 8 节。填写要点：

| 小节 | 怎么填 |
|------|--------|
| 时间范围 | Week 3 实际跨了 2026-07-09（Day 15）到 2026-09-03（Day 21）。真实节奏就是这样，不用美化成 7 天 |
| 本周目标 | Web 框架 + 数据库 + 分层，从「能返回假数据的 handler」到「数据真的落库的三层服务」 |
| 已完成 Issue | #15 ~ #21 |
| 本周代码产出 | 按包写：`internal/handler`、`internal/service`、`internal/repository`、`internal/middleware`、`internal/config`、`internal/model`、`pkg/apperror`、`pkg/response`、`deployments/` |
| 掌握的概念 | 挑真正理解了的写，比如接口归属、错误包装 + `errors.As` 统一映射、middleware 执行顺序、nil slice 序列化成 `null` |
| 仍然薄弱的点 | 诚实写。参考：GORM 的关联和预加载没碰过；事务只写了一个玩具例子；没做过分页；JWT 完全空白；集成测试靠环境变量门控，平时不跑 |
| 代码 Review 结论 | 汇总本周几次 review 里反复出现的问题：脚手架注释没清理、import 分组、`gofmt` 不重排分组、测试里 panic 掩盖其他失败 |
| 下周优先级 | Day 22 数据库设计、Day 23 用户与认证、Day 24 完整 CRUD、Day 25 JWT |

另外有一个计划层面的问题要一起处理：`docs/weekly/week-02.md` 还是空模板。`GO_LEARN_PLAN.md` 里 Day 7 明确写了产出 `week-01.md`，但 Day 14 是「综合练习」，产出指向 `week02-core/day14-summary/`，**没有任何一天被安排写 week-02.md**。这是计划的漏洞，不是漏做的任务。今天顺手补上比较合适，内容从 Day 8 ~ Day 14 的日志里回收。

---

## 今日代码结构

改动的文件：

```
internal/middleware/recovery.go   # 新增，Recovery middleware
pkg/apperror/error.go             # 新增 NotFound 构造函数
internal/handler/notfound.go      # 新增，NotFoundHandler
internal/handler/todo.go          # 绑定错误改成固定文案 + 打日志
cmd/server/main.go                # 挂 Recovery；NoRoute 改为 handler.NotFoundHandler
go.mod / go.sum                   # go mod tidy
```

加测试：

```
internal/middleware/recovery_test.go   # panic → 500 + 统一信封 + 不中断进程
internal/handler/notfound_test.go      # 404 走完整 router，断言错误信封形状
internal/handler/todo_test.go          # 补 message 断言（现有断言只查 code，改完照样过）
pkg/apperror/error_test.go             # 断言 NotFound 的 code / status
```

文档动五个：`docs/api/todo-api.md`、`docs/architecture/todo-api.md`、`docs/weekly/week-03.md`、`docs/weekly/week-02.md`、`README.md`（补 `make run` 的 MySQL 前置和 `curl` 示例）。

---

## 今日验收标准

1. 端到端跑通：Docker + `make run` + 两条 `curl`，POST 拿到非 0 `id`，重启服务后 GET 数据还在。
2. handler 里 panic 时，客户端收到 500 + `{"error":{"code":"INTERNAL_ERROR",...}}`，且日志里有一条带 request ID 的记录。
3. `pkg/apperror` 有 `NotFound`，返回 `NOT_FOUND` / 404；访问任意不存在的路径响应形状与其他错误一致。
4. 绑定失败的响应里不含 `CreateTodoRequest` 之类的内部类型名。
5. `go mod tidy -diff` 无输出。
6. `docs/api/todo-api.md` 区分「已实现」和「计划中」，已实现部分的每个例子都能跑通。
7. `docs/architecture/todo-api.md` 写上依赖方向、接口归属、依赖组装位置。
8. `docs/weekly/week-03.md` 8 节填完。
9. `make fmt`、`make test`、`make vet` 全部通过。

## 可选挑战题

1. 给架构文档画一张 ASCII 依赖图，标出哪些方向是禁止的。
2. 把 `writeError` 的兜底分支从手写字面量改成 `apperror.Internal("internal server error")`，消掉重复的 `"INTERNAL_ERROR"` 字符串。
3. 在 `docs/api/todo-api.md` 里给每个「计划中」的接口标上负责的 Day，形成一张待办地图。

## 今天最容易踩的坑

1. **只跑 `make test` 就认为雏形能用。** 所有单元测试都用替身，一行 SQL 都没执行过。必须真的连库、真的重启。
2. **先跑集成测试再做手工验证。** `internal/repository/todo_test.go` 的 teardown 会 `DROP TABLE IF EXISTS todos`（`:38`），每个用例开头还会 `DELETE FROM todos`（`:49`），会把你刚 POST 进去的数据连表一起删掉，然后你会以为持久化坏了。顺序反过来：先手工验证，再跑集成测试。

   补充（实现后核实）：这条只在 `RUN_INTEGRATION_TESTS=true` 时成立 —— `TestMain`（`:18`）没看到这个环境变量就直接 `m.Run()`，`testDB` 保持 nil，所有用例 `t.Skip`，既不建表也不删表。所以裸 `make test` / `go test ./...` 不会碰数据库，代价是**它也没有跑任何真实 SQL**，看到 `ok internal/repository` 别以为数据库路径被覆盖了。
3. **把 `Recovery` 挂在 `Logging` 外层。** 直觉会让你这么写，但那样访问日志里依然不会出现这次请求 —— 见改动二的实测。正确顺序是 `RequestID, Logging, Recovery, CORS, AuthPlaceholder`。
4. **在 `Recovery` 里返回 panic 的原文。** `recover()` 拿到的值可能带内部路径、SQL 片段，和改动三是同一类问题：原文进日志，对外只给固定文案。
5. **改文档时只删不标。** 把未实现的接口直接删掉，下周又要凭记忆重写。标成「计划中 + Day NN」比删除有用。
6. **`NotFoundHandler` 放错包。** `writeError` 是 `internal/handler` 的包级私有函数，`NotFoundHandler` 必须在同一个包里才能调用；放到 `main` 里就得把错误信封逻辑复制一遍，那就白改了。
7. **测 404 时只起裸 `gin.New()`。** 那样测不到「404 也带 request ID」这条结论。要按 `main.go` 的顺序挂上全局 middleware 再测。
8. **周复盘写成成果汇报。** 「仍然薄弱的点」是这份文档里唯一有长期价值的一节，写虚了就等于没写。




