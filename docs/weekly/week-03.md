# Week 03 复盘：Gin, GORM, Project Architecture

## 时间范围

2026-07-09 ~ 2026-09-08

日历上跨了两个月，实际只有 7 个学习日（Day 15 和 Day 16 是同一天，Day 18 到 Day 21 之间各有大段空档）。这本身是本周最该记下的一条：断档期间上下文全靠 `docs/daily/` 找回来，日志写得细的那几天恢复得快，写「待记录」的那几天基本等于没写。

## 本周目标

- 从 `net/http` 切到 Gin，建立路由组和统一 JSON 响应。
- 打通 MySQL：连接、迁移、CRUD，再从 `database/sql` 升级到 GORM。
- 补上中间件层：请求 ID、日志、CORS、认证占位。
- 把代码重构成 handler / service / repository 三层，依赖只在 `main` 组装。
- 让 TODO API 雏形真正跑起来，而不只是单元测试通过。

## 已完成 Issue

- Day 15: Add Gin and route groups（Gin 框架和路由组）
- Day 16: Add TODO request binding and JSON responses（TODO 请求绑定和 JSON 响应）
- Day 17: Add MySQL connection and basic CRUD（MySQL 连接和基础 CRUD）
- Day 18: Add GORM models and repository implementation（GORM 模型和 repository）
- Day 19: Add logging/auth/CORS middleware（中间件层）
- Day 20: Refactor into handler/service/repository layers（分层架构）
- Day 21: Complete the API service skeleton and Week 3 review（API 骨架收尾与周复盘）

## 本周代码产出

- Day 15：引入 Gin 依赖；`internal/handler/health.go` 用 Gin 返回 JSON；`cmd/server/main.go` 换成 Gin 路由器并建 `/api/v1` 路由组。测试覆盖成功响应和方法不允许。
- Day 16：`internal/handler/todo.go` 实现 `TodoHandler` / `CreateTodo` / `ListTodos`，`c.ShouldBindJSON` + `binding:"required"` 做绑定校验，创建返回 201，列表返回 `[]Todo{}`。测试覆盖成功创建、无效 JSON、缺必填字段、空列表。
- Day 17：`deployments/docker-compose.yml`（MySQL 8.0）、`deployments/migrations/001_create_todos.sql`、`internal/config/database.go`（DSN + 连接）、`internal/model/todo.go`、`internal/repository/todo.go`（`database/sql` 版 CRUD），并在容器内的 MySQL 上跑通集成测试。
- Day 18：模型和 repository 整体迁到 GORM（结构体 tag、CRUD、`db.Transaction` 版 `CreateAndMarkDone`），集成测试改用 `AutoMigrate` 建表。
- Day 19：新增 `internal/middleware/` 四个文件 —— `request_id.go`、`logging.go`、`cors.go`、`auth.go` —— 加 `middleware_test.go`，并在 `main.go` 挂载。
- Day 20：新增 `internal/service/todo.go`（title 校验、错误包成 `apperror`、空切片归一化）+ 7 个用例的单元测试（内存假 repository）；新增 `internal/config/gorm.go`（`DefaultDSN` / `ConnectGorm`）；handler 改为通过接口依赖 service，收拢出唯一错误出口 `writeError`；`main.go` 完成 repo → service → handler 的装配。
- Day 21：`internal/middleware/recovery.go`（panic → 500 + 统一信封）、`internal/handler/notfound.go`（`NoRoute` 走统一信封）、`pkg/apperror` 的 `NotFound` 及其测试；修掉绑定错误原文外泄；`go mod tidy`；产出 `docs/issues-backlog.md`（24 条全量审计）并重写 API / 架构 / README 三份文档。

## 掌握的概念

### Gin

- `gin.New()` 创建空路由器，不附加任何中间件；`gin.Default()` 会自动附加 `Logger` 和 `Recovery`。本项目用 `gin.New()`，所以 Recovery 需要自己补 —— 这个漏洞直到 Day 21 才被发现并修掉。
- `router.Group("/api/v1")` 让组内路由共享前缀，好处是不用重复写前缀、可统一挂中间件、便于版本管理。
- `gin.Context` 封装请求和响应；`c.JSON(code, obj)` 一次做三件事：设状态码、设 `Content-Type: application/json`、序列化写响应体。
- `gin.H` 是 `map[string]any` 的别名，用于快速构造 JSON，不必定义结构体。
- 测试用 `gin.SetMode(gin.TestMode)` 关掉路由注册日志，再用 `router.ServeHTTP(recorder, req)` 模拟请求，不需要监听端口。

### 请求绑定与响应

- `BindJSON` 失败时自动返回 400 并终止请求，错误格式无法自定义；`ShouldBindJSON` 只返回错误，由调用方决定怎么响应。要统一错误信封就必须用后者。
- `binding:"required"` 由 Gin 的 validator 校验；去掉后字段变可选，缺失时取零值。
- nil slice 序列化成 `null`，`[]Todo{}` 序列化成 `[]`。前端遍历 `null` 会报错，所以列表接口一律用 `make([]Todo, 0, len(x))` 起手。
- POST 创建资源返回 201 Created，语义比 200 精确。

### database/sql（Day 17，已被 GORM 取代）

- `sql.Open` 只创建 `*sql.DB` 和连接池对象，不保证连通；`db.Ping()` 才真实发起连接。
- DSN 里 `parseTime=True` 让驱动把时间列转成 `time.Time`，去掉会拿到 `[]byte`/`string`。
- `sql.ErrNoRows` 是「查询成功但没数据」，不是故障，repository 层应转成 `nil, nil`。
- `rows.Close()` 必须 `defer`，否则连接泄漏、池耗尽、后续查询排队超时；遍历完还要检查 `rows.Err()`。
- 用占位符让驱动绑定参数，不手动拼 SQL，避免注入。

### GORM

- tag 语义：`primaryKey` 主键、`size:255` 列长度、`not null` 非空、`index` 索引；`gorm.DeletedAt` 开启软删除。
- `Create` 会把数据库生成的自增 ID 通过传入的结构体指针写回调用方。
- `AutoMigrate` 只做「建不存在的表、加新字段、加缺失索引」，不处理删除、重命名、复杂类型变更和数据迁移，也没有版本记录和回滚，**不能替代正式迁移工具**。
- `gorm.ErrRecordNotFound` 是可预期的业务分支，要和连接/语法/权限类错误区分开。
- `db.Transaction` 的回调返回 `nil` 提交、返回非 nil 回滚 —— 内部错误必须 `return` 出去，否则事务被当成成功提交。

### 中间件

- 中间件放请求级别的横切逻辑：日志、跨域、链路追踪、panic 恢复。典型结构是「`c.Next()` 之前准备，之后收尾」。
- 状态码和耗时必须在 `c.Next()` 之后读，之前拿 `c.Writer.Status()` 只会得到默认 200；但**请求路径在 `c.Next()` 之前就能取到**，它来自 `c.Request.URL`。
- 请求 ID 要同时写进 context 和响应头，少写哪一边链路都会断。Go 的 `http.Header` 会把名字规范化成 `X-Request-Id`，写断言时别按字面比较。
- CORS 预检只在非简单请求时触发（自定义头、`PUT`/`DELETE`、非表单 Content-Type），返回 204 最合适，而且必须排在鉴权之前，否则浏览器看到预检被 401 拦下会报跨域错误。
- **挂载顺序不是随意的**（Day 21 实测）：`Recovery` 必须在 `Logging` 之内，否则 panic 被兜住时控制流回不到 `Logging` 的 `c.Next()` 之后，那次请求没有访问日志；`RequestID` 在最外层，`Recovery` 才能在日志里带上它；`CORS` 在 `Recovery` 之内，500 才保留跨域头。
- `Recovery` 里的 `Abort` 是**必需的，不是惯例** —— 但只在特定场景下体现出来。`c.Next()` 是 `c.index++` 加循环，panic 解栈后控制流回到外层的 `Next()` 循环并继续 `index++`：
  - panic 发生在某个中间件**调用 `c.Next()` 之前**：恢复后 `index++` 正好落在下一个 handler 上，后续中间件和业务 handler 会照常执行，响应变成 `{"code":"INTERNAL_ERROR"}route ok` 这种拼接的坏 JSON。`Abort` 把 `index` 设成 63 才拦住。
  - panic 发生在业务 handler（链的最后一环）：`index` 已到末尾，加不加 `Abort` 都一样。
  - panic 发生在某个中间件 `c.Next()` **之后**的收尾代码里：业务 handler 已经写完 200 和 body，`Recovery` 只能追加，**状态码停在 200**，gin 打 `Headers were already written`。这种 `Recovery` 兜不住，`Abort` 也救不了。
- `router.NoRoute` 注册的 handler **会**经过全局 `router.Use` 的中间件，所以 404 上一样有请求 ID 和跨域头。

### 分层架构（本周最重要的一块）

- 四层职责：`main` 组装、handler 管 HTTP、service 管业务、repository 管数据库。
- **接口定义在调用方，且只声明用得到的方法。** 这和 Java 相反。`repository.TodoRepository` 有 6 个方法，service 声明的接口只列 2 个。如果接口定义在 repository 包，service 就得 import repository 才能拿到它，依赖又绑回具体实现了。
- Go 不需要 `implements`，方法集覆盖即隐式满足。
- 错误在 service 层就包成带状态码的 `apperror`，handler 靠 `errors.As` + `appErr.StatusCode` 统一映射，不做 if-else 判断类型。直接 return gorm 错误的后果有两个：参数错误和资源不存在全变成 500（状态码语义丢失），以及 gorm 的错误文本（可能含表名、SQL 片段）泄漏给客户端。
- 边界可以用 import 检查：service 里不出现 `gin` 也不出现 `gorm`，handler 里不出现 `gorm`。
- 依赖只在 `main` 组装，各层都不自己 new 下游 —— handler 自己 new service 就意味着依赖具体类型而非接口，测试时无法注入假实现。
- 引入 service 层的真正理由是「业务规则无处安放」，不是目录看起来规范。分层最直接的回报是可测性：service 用假 repository、handler 用假 service，两层都不碰数据库。

### 安全与工程实践

- 面向客户端的错误文案要固定，validator / json / gorm 的原始错误只写日志，按 `request_id` 关联即可回溯。原文里会有内部结构体名、Go 类型名、表名。
- panic 值和调用栈同理：只写日志，不返回。
- `make fmt test vet` 不需要 MySQL（单元测试全用替身），数据库只在 `make run` 和集成测试时需要。
- repository 的集成测试由 `RUN_INTEGRATION_TESTS=true` 门控：不设这个变量时 `TestMain` 直接 `m.Run()`，`testDB` 为 nil，所有用例 `t.Skip`。所以 `ok internal/repository` **不代表真实 SQL 跑过了**。设了变量时 teardown 会 `DROP TABLE`、每个用例开头 `DELETE FROM todos`，此时手工验证必须排在它之前。
- `gofmt` 只在 import 组内排序，不会重新分组，需要 `goimports`。
- 用 `curl -D-` 看响应头，比只看单元测试更能确认中间件的实际效果。

## 仍然薄弱的点

- **写完代码不等于跑通。** Week 3 的检查点是「TODO API 雏形可以运行」，但 Day 20 收尾时 colima 没启动，端到端 curl 一直没做，直到 Day 21 才第一次真实跑通。所有单元测试都用替身，一行 SQL 都没经过 —— 绿色的测试给了「已经能跑」的错觉。
- **文档不同步会反噬。** 一周之内 API 文档漂到描述 7 个不存在的接口，架构文档还写着「初期可使用内存实现」。Day 21 的审计里文档类问题占了 5 条，其中 3 条是 P1。
- **自己不用的东西写不对。** `pkg/response` 的 `GinJSON`/`GinError` 用 `c any` + 内联接口断言，断言失败静默返回 —— 零调用方，所以这个错误模式没人发现（backlog B2）。
- 连接池参数 `MaxOpenConns` / `MaxIdleConns` 设了但从未生效，`ConnectGorm` 没有通过 `db.DB()` 去配置池（backlog A4，Day 26）。
- 迁移脚本和模型不一致，且从未被执行 —— 实际建表靠 `AutoMigrate`，项目里有两套 schema 来源，一套是死的（backlog C2，Day 22）。
- `AuthPlaceholder` 挂在全局，Day 25 真正实现 JWT 时会把 `/api/v1/health` 一起拦掉（backlog A6）。
- CORS 反射任意 `Origin`，加认证之前必须先收紧成白名单（backlog A5）。
- repository 测试的数据隔离和清理还没解决：硬编码 DSN，teardown 直接 `DROP TABLE`（backlog C5，Day 22）。
- `make test` 没有 `-count=1`，Day 20 已经被测试缓存骗过一次（backlog C3）。
- 「答疑记录」在 Day 15 / 16 / 17 / 18 全写「待记录」—— 教案阶段的提问环节这四天是空的。
- 事务只有 `CreateAndMarkDone` 一个骨架示例，没有真实业务场景。

## 典型错误与修正

- Day 16：`pkg/response` 原本是给 `net/http` 写的，Gin 里不能直接用 → 改为 `c.JSON()` 搭配 `response.Body` 结构体。
- Day 17：本机没装 `mysql` 客户端 → 改用 `docker exec` 在容器内执行迁移。
- Day 17：怀疑 `docker-compose.yml` 的 `mysql_data:` 写错 → 确认那是正确的命名卷声明。
- Day 18：把 `gorm.io/driver/mysql` 导进了 repository 文件 → 驱动只在连接层需要，放这里会变成未使用导入。
- Day 18：GORM 查询未命中时打印 `record not found` 日志 → 属于预期行为，在 repository 层单独处理成 `nil, nil`。
- Day 19（小测）：以为请求路径也要等 `c.Next()` 之后才能取 → 路径来自 `c.Request.URL`，前面就有；正确写法是 `start := time.Now()` 在前，`status` 和 `latency` 的读取在后。
- Day 19：`RequestID` 的兜底值一开始写得不合适 → 改为固定字符串 `request-id`。
- Day 19：`request_id.go` 误加了 `net/http` 导入 → 移除。
- Day 19（流程）：middleware 由 AI 直接写完，跳过了「空骨架由本人填充」这一步，缺了代码巩固环节 → Day 20 起恢复正确流程。
- Day 19（流程）：小测试未实际作答就提交并关闭了 Issue → 事后补答补批改。
- Day 20（小测）：只说出错误包装的方向，漏了两个具体后果 → 补上「状态码语义丢失，全变 500」和「gorm 错误文本泄漏表名/SQL 片段」。
- Day 20：`strings` 排到了项目包之后 → `gofmt` 不重新分组，需要 `goimports`。
- Day 20：脚手架 `// TODO: implement` 注释未清理 → 会让人误以为代码未完成；`internal/repository/todo.go` 从 Day 18 起就有同样遗留，一并清掉。
- **Day 21：同一个脚手架注释问题在 `pkg/apperror/error.go` 上复发。** 连续两天出现，说明这不是偶发，填完实现要 `grep -rn "TODO: implement"` 扫一遍。
- Day 21（我写教案时的错误，两轮）：先发布「`Recovery` 必须挂在 `Logging` 之前」，写测试时实测是反的，改掉。又发布「`c.JSON` 不加 `Abort` 后面的 handler 还会执行」→ 实测一种场景后改成「`Abort` 只是惯例」→ **这个更正本身也是错的**，被小测第 2 题的作答纠正：panic 发生在中间件调用 `c.Next()` 之前时，不 `Abort` 后续 handler 真的会跑，响应会拼成坏 JSON。根因是我只测了一种 panic 位置就推广了结论。
- Day 21（验证时的坑）：`:8080` 上还挂着 9 月 3 日 `make run` 留下的旧进程，新二进制 `bind: address already in use` 直接退出，第一轮 curl 全打在旧代码上，看起来像修的东西没生效。验证前先确认端口上跑的是哪个二进制。

## 代码 Review 结论

- Day 15：`gin.New()` + `Group("/api/v1")` 的结构清晰，测试用 `TestMode` + `router.ServeHTTP`，四项命令通过。
- Day 16：`TodoHandler` 用结构体方法，为后续注入依赖留好了位置；`ShouldBindJSON` 失败手动返回 400；统一用 `response.Body`；创建返回 201；空列表用 `[]Todo{}`。
- Day 17：DSN 三个参数齐全；`sql.Open` + `Ping` 验证连通；全程参数化 SQL；`FindByID` 处理 `sql.ErrNoRows`；`FindAll` 检查 `rows.Err()`；集成测试跑在真实 MySQL 上。
- Day 18：GORM tag 定义完整（含软删除）；`Create` 回填 ID；`CreateAndMarkDone` 用 `db.Transaction` 包裹多步写。
- Day 19：`RequestID` 优先用客户端传入值；`Logging` 在 `c.Next()` 后记录，不打断响应；`CORS` 对 `OPTIONS` 返回 204；`AuthPlaceholder` 不影响公开接口。
- Day 20：分层边界正确（service 无 `gin`/`gorm`，handler 无 `gorm`）；接口定义在调用方且只声明 2 个方法；校验失败即返回、不白跑数据库（用 `createCalls` 断言）；错误统一包成 `apperror` + `errors.As` 映射；依赖只在 `main` 组装。
- Day 21：`Recovery` 的实现处理了「panic 值只写日志、不回客户端」这个关键点，并用 request ID 把 panic 日志和访问日志串起来；`NotFoundHandler` 放在 `internal/handler` 包内以复用 unexported 的 `writeError`，是对包边界的正确处理；`internal/middleware` 拿不到 `writeError`，改为用 `pkg/response` 重建同形状信封并写明了原因。唯一的返工是残留的脚手架注释。

## 验证命令

```bash
# 每日固定
make fmt
make test
make vet

# 数据库相关（Day 17、18、21）
docker compose -f deployments/docker-compose.yml up -d
docker exec -i go_learning_mysql mysql -uroot -ppassword go_learning < deployments/migrations/001_create_todos.sql
RUN_INTEGRATION_TESTS=true go test ./internal/repository -v

# 绕开测试缓存（Day 20 起）
go test -count=1 ./...

# 端到端（Day 19 看响应头，Day 21 第一次完整跑通）
make run
curl -sS -D - http://127.0.0.1:8080/api/v1/health
curl -sS -X POST http://127.0.0.1:8080/api/v1/todos -H 'Content-Type: application/json' -d '{"title":"learn go"}'
curl -sS http://127.0.0.1:8080/api/v1/todos

# 模块整洁度（Day 21）
go mod tidy -diff
```

## 下周优先级

1. **Day 22 先做决策再写代码**：`Done bool` 还是 `Status string`（backlog D3），以及 `AutoMigrate` 和迁移脚本以哪套为准（C2）。这两条不定下来，后面每加一个字段都要还一次债。
2. 补完 TODO CRUD：`GET/PUT/DELETE /api/v1/todos/{id}` 加分页筛选（Day 24），顺手接上 repository 里那四个还没调用方的方法。
3. 用户与认证：User 模型 + 密码哈希（Day 23）、JWT（Day 25）。落地前必须先处理 A5（CORS 白名单）和 A6（`AuthPlaceholder` 从全局挪到受保护路由组）。
4. 每天改完代码顺手同步 `docs/api/todo-api.md` 的「已实现」一节，别再攒到周末一次性还债。
5. 每条修掉的问题在 `docs/issues-backlog.md` 里改状态，不删行。


