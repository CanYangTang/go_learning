# 项目问题清单

Day 21（2026-09-03）对整个项目做了一次全量审计，本文件是结果。每条都对着代码验证过，证据列写的是验证方式或文件位置。

维护方式：修掉一条就把「状态」改成 `已修 (Day NN)`，不要删行 —— 保留记录比保持清单干净有用。新发现的问题追加到对应小节，编号不复用。

优先级：`P0` 会导致线上事故或数据/信息泄漏；`P1` 破坏对外契约；`P2` 埋雷，下一步开发前必须处理；`P3` 一致性和整洁度。

## 状态总览

| 编号 | 问题 | 优先级 | 计划 Day | 状态 |
|------|------|--------|----------|------|
| A1 | 缺少 Recovery middleware，panic 不返回 500 | P0 | 21 | 已修 (Day 21) |
| A2 | `NoRoute` 的 404 响应形状与统一信封不一致 | P1 | 21 | 已修 (Day 21) |
| A3 | 绑定失败把 validator 原文和 Go 类型名回给客户端 | P1 | 21 | 已修 (Day 21) |
| A4 | 连接池参数 `MaxOpenConns`/`MaxIdleConns` 从未生效 | P2 | 26 | 待修 |
| A5 | CORS 反射任意 `Origin`，无白名单 | P2 | 25 前置 | 已修 (Day 25) |
| A6 | `AuthPlaceholder` 挂在全局，JWT 落地会保护掉 health | P2 | 25 | 已修 (Day 25) |
| A7 | health 响应不走统一信封（项目里共 4 种响应形状） | P3 | 21 决策 | 已决策 (Day 21) |
| B1 | `database/sql` 整条路径已成死代码 | P3 | 26 | 待修 |
| B2 | `pkg/response` 四个函数零调用，`GinJSON`/`GinError` 模式错误 | P3 | 24 | 已修 (Day 26) |
| B3 | repository 四个方法无调用方 | P3 | 24 | 部分已修 (Day 24) |
| B4 | `health_test.go` 手写了 `strings.Contains` | P3 | 27 | 已修 (Day 26) |
| C1 | `go.mod` 未 tidy，四个直接依赖被标成 `// indirect` | P2 | 21 | 已修 (Day 21) |
| C2 | 迁移脚本与模型不一致，且从未被执行 | P2 | 22 | 已修 (Day 22) |
| C3 | `make test` 无 `-count=1`；`fmt` 管不了 import 分组；无 lint | P2 | 27 | 待修 |
| C4 | `pkg/apperror`、`pkg/response` 无测试 | P3 | 21 / 24 | 已修 (Day 21 apperror / Day 26 response) |
| C5 | 集成测试硬编码 DSN，teardown 会 `DROP TABLE` | P2 | 22 | 待修 |
| C6 | `week02-core` 里残留 8 处 `// TODO: implement` 脚手架注释 | P3 | 27 | 待修 |
| D1 | `docs/api/todo-api.md` 描述了 7 个不存在的接口，6 处形状不符 | P1 | 21 | 已修 (Day 21) |
| D2 | `docs/architecture/todo-api.md` 三处过时描述 | P1 | 21 | 已修 (Day 21) |
| D3 | 设计冲突：文档 `Status string` vs 代码 `Done bool` | P2 | 22 决策 | 已决策 (Day 22) |
| D4 | README 的 `make run` 已经跑不起来，且无 endpoint 示例 | P1 | 21 | 已修 (Day 21) |
| D5 | `DB_DSN` 零处文档记录，无 `.env.example` | P2 | 26 | 部分已修 (Day 21) |
| E1 | `docs/weekly/week-02.md` 无人负责（计划漏洞） | P3 | 21 | 已修 (Day 21) |
| E2 | 开发库口令写死在源码里 | P2 | 26 | 待修 |
| E3 | 端到端从未真实跑通过 | P1 | 21 | 已验证 (Day 21) |

## A. 代码缺陷

### A1 — 缺少 Recovery middleware（P0）

`cmd/server/main.go:38-39` 用 `gin.New()` 建 router，只挂了 `RequestID`、`Logging`、`CORS`、`AuthPlaceholder`。`gin.Default()` 自带的 `gin.Recovery()` 被一起丢掉了，之后也没补回来。

实测结论（临时测试，已删除）：handler 里 `panic("boom")` 时

- 客户端收到 `Get "...": EOF`，连接被直接丢弃 —— 不是 500，也不是统一错误信封。
- panic 从 `c.Next()` 里穿过 Logging middleware，`c.Next()` 后面的代码不执行，**这次请求不会产生任何访问日志**。
- 进程本身存活，`net/http` 的 per-connection recover 兜住了，只在 stderr 打 `http: panic serving`。

修法：新增 `internal/middleware/recovery.go`，`defer` + `recover()`，命中后记一条带 request ID 的日志，再走统一错误信封返回 500。

挂载位置：必须在 `Logging` **之内**（后面）。审计时我一开始写的是「挂在 Logging 之前」，实测发现是反的 —— Recovery 在外层时 panic 在到达 Logging 的 `c.Next()` 后续代码之前就被截住，那次请求只有 panic 日志、没有访问日志；Logging 在外层时两行日志都有。

### A2 — `NoRoute` 响应形状不一致（P1）

`cmd/server/main.go:48-52` 返回 `{"error": "not found"}`，`error` 是字符串；其他所有错误经 `internal/handler/todo.go:87` 的 `writeError` 返回 `{"error": {"code", "message"}}`，`error` 是对象。客户端写统一错误解析后，在任何拼错的 URL 上都会失效 —— 而拼错 URL 是最容易触发的错误。

详细修法见 `docs/daily/day-21-lesson.md`。

### A3 — 绑定错误原文外泄（P1）

`internal/handler/todo.go:46` 把 `err.Error()` 直接塞进对客户端的响应。实测输出：

```json
{"error":{"code":"VALIDATION_ERROR","message":"Key: 'CreateTodoRequest.Title' Error:Field validation for 'Title' failed on the 'required' tag"}}
{"error":{"code":"VALIDATION_ERROR","message":"json: cannot unmarshal array into Go value of type handler.CreateTodoRequest"}}
```

内部结构体名和 Go 类型名都泄漏了，而且对客户端毫无可用信息。修法：对外返回固定文案（如 `invalid request body`），原始错误只写日志。

### A4 — 连接池参数从未生效（P2）

`internal/config/gorm.go:17-18` 设了 `MaxOpenConns: 10, MaxIdleConns: 5`，但这个 `DatabaseConfig` 在这里只被调用了 `.DSN()`，而 `DSN()` 只拼连接字符串。两个字段等于注释。`ConnectGorm` 也没有通过 `db.DB()` 拿 `*sql.DB` 去设置池。读代码的人会以为已经限流了。

### A5 — CORS 反射任意 Origin（P2）

`internal/middleware/cors.go:12-17` 把客户端传来的 `Origin` 原样写回 `Access-Control-Allow-Origin`。当前所有接口都是公开的，实际影响有限；但 Day 25 加 JWT 之后，这等于宣布任何站点都可以跨域读取受保护接口的响应。**加认证之前必须先收紧成白名单。** 顺带缺 `Access-Control-Max-Age`，预检结果不被浏览器缓存。

### A6 — `AuthPlaceholder` 挂在全局（P2）

`cmd/server/main.go:39` 把它放进了 `router.Use`。真正实现 JWT 校验时会连 `/api/v1/health` 一起拦掉。Day 19 的记录里提到过这个边界，但代码里没有任何提示。Day 25 要改成挂在需要保护的路由组上。

### A7 — health 不走统一信封（P3，待决策）

`internal/handler/health.go:11` 返回裸对象 `{status, service, version}`。加上成功信封、错误信封、`NoRoute` 的字符串形状，项目里一共有 4 种响应形状。health 保持裸对象是合理选择（探针通常直接读 `status`），但需要在 API 文档里写明这是有意的例外，而不是漏改。

## B. 死代码

### B1 — `database/sql` 路径已废弃（P3）

`internal/config/database.go` 里的 `Connect(cfg *DatabaseConfig) (*sql.DB, error)` 零调用方（`grep` 确认）。Day 17 写的 `database/sql` 连接方式在 Day 20 换成 GORM 后就废了，`DatabaseConfig` 现在只剩 `DSN()` 被 `DefaultDSN()` 借用。要么删掉 `Connect` 和那个 `_ "github.com/go-sql-driver/mysql"` 空导入，要么在注释里写明它是 Day 17 的学习留档。

### B2 — `pkg/response` 四个函数零调用（P3）

`JSON`、`Error`、`GinJSON`、`GinError` 全部没有调用方（`grep` 确认），只有 `Body`、`ErrorBody`、`ErrorPayload` 三个类型在用。

`GinJSON`/`GinError` 还是个应该删掉的反面模式：参数写成 `c any`，函数体里定义内联接口再做类型断言，断言失败就静默返回、什么都不写。真要接 Gin 就直接依赖 `*gin.Context`；这里绕开依赖的代价是丢掉了编译期检查。

### B3 — repository 四个方法无调用方（P3，观察）

`FindByID`、`Update`、`Delete` 在 Day 24 做完整 CRUD 时会接上，可以接受。`CreateAndMarkDone` 是 Day 18 的事务练习产物，之后大概不会有调用方 —— 到 Day 24 时决定是保留为示例还是删掉。

### B4 — 手写的 `strings.Contains`（P3）

`internal/handler/health_test.go:68-80` 有一对 `containsString` / `findSubstring`，逐字节实现了 `strings.Contains`。同包的 `todo_test.go` 用的是标准库。Day 15 的遗留，删掉换成 `strings.Contains` 即可。

## C. 工程配置

### C1 — `go.mod` 未 tidy（P2）

`go mod tidy -diff` 确认：`github.com/gin-gonic/gin`、`github.com/go-sql-driver/mysql`、`gorm.io/driver/mysql`、`gorm.io/gorm` 四个**直接**依赖全被标成了 `// indirect`。`go.sum` 也缺几行。一条 `go mod tidy` 解决。

### C2 — 迁移脚本与模型不一致（P2，已修 Day 22）

`deployments/migrations/001_create_todos.sql` 只有 `id/title/done/created_at/updated_at`，缺 `deleted_at`；而 `internal/model/todo.go:16` 用了 `gorm.DeletedAt`。

更根本的问题是：实际建表靠 `cmd/server/main.go:30` 的 `AutoMigrate`，这个 SQL 文件从来没有被执行过。项目曾经有两套 schema 来源，其中一套是死的。

**决策（Day 22）**：`model.Todo` 的 struct tag + `AutoMigrate` 是唯一 schema 来源，`001_create_todos.sql` 已删除。已知代价写进了 `docs/architecture/todo-api.md`「Day 22 设计决策」：不处理删除列/重命名/类型变更，无版本记录和回滚。

### C3 — 构建命令的三个缺口（P2）

- `make test` 是裸 `go test ./...`，Day 20 已经被测试缓存骗过一次（`make fmt` 改完文件后仍报 `(cached)`）。应加 `-count=1`，或另开一个 `make test-fresh`。
- `make fmt` 是 `go fmt`，只在 import 组内排序，不会重新分组 —— Day 20 那个 `strings` 排到项目包后面的问题会复发。需要 `goimports`。
- 没有 lint target。

### C4 — 两个包无测试（P3）

`pkg/apperror`、`pkg/response` 都没有测试文件。`apperror` 今天要加 `NotFound`，正好一起补上 code/status 的断言。

### C5 — 集成测试的两个隐患（P2）

`internal/repository/todo_test.go:22` 硬编码 DSN，不读 `DB_DSN`，和 `config.DefaultDSN()` 重复了一份。

teardown 里执行 `DROP TABLE IF EXISTS todos`（`:38`），每个用例开头还 `DELETE FROM todos`（`:49`）：跑一次集成测试会把本地库的 todos 表整个删掉。做端到端手工验证时要注意顺序 —— 先验证，再跑集成测试，否则数据会被清掉，让人误以为持久化失败。

补充（Day 21 核实）：以上只在 `RUN_INTEGRATION_TESTS=true` 时成立。`TestMain`（`:18`）看不到这个变量就直接 `m.Run()`，`testDB` 保持 nil，八个用例全部 `t.Skip`。**这带来一个更隐蔽的问题：裸 `go test ./...` 输出的 `ok internal/repository` 完全不代表数据库路径被覆盖过**，它一行 SQL 都没执行。Day 22 一并处理：DSN 读 `DB_DSN`、teardown 别 `DROP TABLE`（用独立测试库或 `TRUNCATE`），并让跳过时的输出更显眼。

### C6 — `week02-core` 残留脚手架注释（P3，Day 21 新增）

`grep -rn "TODO: implement" --include="*.go" .` 命中 8 处，全在 `week02-core`：`day12-json/handler.go:21`、`day13-files/reader.go:10,20`、`day13-files/writer.go:10,20`、`day14-summary/handler.go:10`、`day14-summary/task.go:29,41`。这些函数**都已经实现好了**，注释只是没删，读代码的人会以为是空壳。

这是同一类问题的第三次出现：Day 18 的 `internal/repository/todo.go`（Day 20 清掉）、Day 20 发现的、Day 21 的 `pkg/apperror/error.go`。填完实现之后 `grep` 一遍应该变成收尾动作的一部分，Day 27 做工程收尾时一并清掉。

## D. 文档
### D1 — `docs/api/todo-api.md` 大面积漂移（P1）

用现在时描述了 7 个不存在的接口，另有 6 处字段和响应形状与代码不符。逐条对照表见 `docs/daily/day-21-lesson.md` 的「第二件事：文档漂移」。

判断标准：文档里每个 `curl` 例子现在跑一遍，都应该得到文档里写的结果。做不到的移到「计划中」并标上 Day。

### D2 — `docs/architecture/todo-api.md` 过时（P1）

- `:48` 「初期可使用内存实现」—— 从来没有内存实现，现在是 GORM + MySQL。
- `:65-73` User 模型 —— 还不存在（Day 23）。
- `:75-86` TODO 表缺 `Done`、`DeletedAt`。

要补的是 Day 20 定下但一个字都没记录的三条设计决策：依赖方向单向、接口定义在调用方、依赖只在 `main` 组装。这三条是整个项目可测性的来源，不写进文档，下次重构就会被破坏。

### D3 — 设计冲突：`Status string` vs `Done bool`（P2，已决策 Day 22）

架构文档规划 TODO 状态是 `Status string`（pending/done），代码已经落成了 `Done bool`。这不是漂移，是一个之前没做的决定。

**决策（Day 22）**：保留 `Done bool`。理由是 YAGNI —— 三周的实际使用和 API 文档规划的接口里都没出现过需要第三种状态的场景，为假设的未来需求现在承担迁移成本和字段类型退化不划算。理由和两条路线的代价对比记在 `docs/architecture/todo-api.md`「Day 22 设计决策」。

### D4 — README 的启动说明已失效（P1）

`README.md:72` 把 `make run` 和 `make fmt` 并列在「常用命令」里，但 Day 20 之后它连不上 MySQL 就 `log.Fatal`，需要先 `docker compose -f deployments/docker-compose.yml up -d`。README 里也没有任何 endpoint 或 `curl` 示例。

Day 28 的验收标准是「项目可按 README 启动」，按现在的 README 做不到。

### D5 — `DB_DSN` 无任何记录（P2）

它是目前唯一的配置开关，只存在于 `cmd/server/main.go:21`，README、API 文档、架构文档都没提，也没有 `.env.example`。

## E. 计划与流程

### E1 — `week-02.md` 无人负责（P3）

`docs/weekly/week-02.md` 是空模板。`GO_LEARN_PLAN.md` 里 Day 7 产出 `week-01.md`、Day 21 产出 `week-03.md`、Day 28 产出 `week-04.md`，而 Day 14 是「综合练习」，产出指向 `week02-core/day14-summary/` —— **没有任何一天负责 week-02.md**。这是计划的漏洞，不是漏做的任务。Day 21 顺手补，内容从 Day 8 ~ Day 14 的日志回收。

### E2 — 开发库口令写死在源码（P2）

`internal/config/gorm.go:15` 的 `Password: "password"` 与 `deployments/docker-compose.yml` 的 `MYSQL_ROOT_PASSWORD` 对应，是本地 Docker 开发库的口令，仓库公开但没有新增暴露面（compose 文件里本来就有）。Day 26 迁到环境变量，`DB_DSN` 已经是现成的覆盖入口。

### E3 — 端到端从未真实跑通（P1）

Week 3 的检查点写了「TODO API 雏形可以运行」，但从来没有被验证过：Day 20 收尾时本机 colima 没启动，端到端只跑到单元测试。而所有单元测试都用替身 —— service 测试用内存假 repository，handler 测试用假 service，**一行 SQL 都没执行过**。

Day 21 的第一件事就是补这个，关键检查是重启服务后数据还在。

## 已修

一条问题修完后保留原小节（便于回看当时的证据），在这里补一行结果。commit 一栏留空是因为记录和修改在同一个 commit 里，用 `git log --oneline docs/issues-backlog.md` 反查即可。

### Day 21（2026-09-08）

| 编号 | 结果 | 落地位置 |
|------|------|----------|
| A1 | 新增 Recovery middleware，panic 返回 500 + 统一信封，panic 值和调用栈只写日志 | `internal/middleware/recovery.go`、`recovery_test.go` |
| A2 | `NoRoute` 改走 `writeError`，返回 `{"error":{"code":"NOT_FOUND",...}}` | `internal/handler/notfound.go`、`notfound_test.go`、`pkg/apperror/error.go` 的 `NotFound` |
| A3 | 绑定失败对外返回固定文案，validator 原文按 `request_id` 写日志 | `internal/handler/todo.go`、`todo_test.go` 的 `TestCreateTodoDoesNotLeakBindErrorDetails` |
| A7 | 决策：health 保持裸对象，已在 API 文档里写明是有意的例外 | `docs/api/todo-api.md`「响应信封」 |
| C1 | `go mod tidy`，四个直接依赖不再是 `// indirect`，`go mod tidy -diff` 干净 | `go.mod`、`go.sum` |
| C4 | `pkg/apperror` 补上测试（code/status 断言 + `errors.As` 可匹配性）。`pkg/response` 仍无测试，留到 B2 一起处理 | `pkg/apperror/error_test.go` |
| D1 | API 文档重写成「已实现 / 计划中」两节，已实现部分的示例全部是从运行中的服务上抓的 | `docs/api/todo-api.md` |
| D2 | 补「当前状态」小节（真实分层、真实 TODO 字段、中间件顺序理由）+ Day 20 的三条设计决策 | `docs/architecture/todo-api.md` |
| D4 | README 新增「本地启动」：MySQL 前置步骤、`DB_DSN` 覆盖方式、五条可直接粘的 `curl` | `README.md` |
| D5 | `DB_DSN` 已在 README 记录（`.env.example` 仍缺，留到 Day 26） | `README.md` |
| E1 | 补写 `week-02.md`，内容从 Day 8 ~ Day 14 的日志回收 | `docs/weekly/week-02.md` |
| E3 | 端到端真实跑通：`docker compose up -d` + 起服务，健康检查 / POST / GET / 400 / 404 / 预检 / request ID 全部实测，**重启服务后数据仍在** | 记录见 `docs/daily/day-21.md` |

顺带发现（Day 21）：`:8080` 上还挂着一个 9 月 3 日 `make run` 留下的旧进程，新编译的二进制起不来（`bind: address already in use`），第一轮 curl 全打在了旧代码上，A3 看起来像没修。教训是验证前先确认端口上跑的是哪个二进制。

### Day 22（2026-09-09）

| 编号 | 结果 | 落地位置 |
|------|------|----------|
| D3 | 决策：保留 `Done bool`，理由是 YAGNI（三周内无第三种状态的真实需求） | `docs/architecture/todo-api.md`「Day 22 设计决策」 |
| C2 | 决策：`AutoMigrate` 为唯一 schema 来源，删除不一致且从未执行的 `001_create_todos.sql` | `docs/architecture/todo-api.md`「Day 22 设计决策」，`deployments/migrations/` 已删除 |

顺带完成：`User` 表设计定案（`Email` 唯一索引、`PasswordHash` 隐藏于 JSON、不带软删除），`UserID` 外键明确留到 Day 25（JWT 落地后才有值可写）。

### Day 24（2026-09-20）

实现 TODO 完整 CRUD：`GET /api/v1/todos/:id`、`PUT /api/v1/todos/:id`、`DELETE /api/v1/todos/:id`（`internal/service/todo.go`、`internal/handler/todo.go`、`cmd/server/main.go`）。`:id` 解析（`strconv.ParseUint`）放在 handler，非法 id 返回 400。PUT 先 `FindByID` 判存在性，避免 `db.Save` 对不存在的 id upsert 出 ghost 记录（端到端实测 `PUT /todos/777` → 404 且未创建记录）。DELETE 采用**选择 B**：`repository.Delete` 签名改为 `(int64, error)` 返回 `RowsAffected`，service 靠 `RowsAffected == 0` 判 404，省去 find-then-delete 的一次查询。

- **B3**：`FindByID`/`Update`/`Delete` 三个方法接上调用方；`CreateAndMarkDone`（Day 18 事务练习）仍无调用方，保留为示例，Day 27 收尾时再决定去留。
- 代码审查修正：初版 `DeleteTodo` 把选择 A（`FindByID`）与选择 B（`RowsAffected`）混用，收敛到选择 B。
- 教案/文档修正：DELETE 成功 body 因 `response.Body.Data` 的 `omitempty` 实际是 `{"message":"ok"}`，非 `{"data":null,...}`。

### Day 25（2026-09-22）

落地 JWT 鉴权（范围决策：**只做鉴权，不做按用户隔离 TODO**，`Todo.UserID` 继续留后续）。

| 编号 | 结果 | 落地位置 |
|------|------|----------|
| A5 | CORS 由「反射任意 Origin」收紧为白名单：仅对 `CORS_ALLOWED_ORIGINS`（默认 `http://localhost:3000`）内的 Origin 回显 ACAO，补上 `Access-Control-Max-Age: 600` 和 `Vary: Origin` | `internal/middleware/cors.go`、`internal/config/env.go` 的 `AllowedOrigins()`、`middleware_test.go` |
| A6 | `AuthPlaceholder` 删除，改为 `RequireAuth(*auth.Manager)` 挂在受保护路由子组 `v1.Group("")` 上；`/health`、`/users/register`、`/users/login` 留在 `v1` 保持公开，5 条 `/todos` 进子组 | `cmd/server/main.go`、`internal/middleware/auth.go` |

新增能力（非 backlog 条目）：

- 新包 `internal/auth`：`Manager.Generate/Parse`（HS256，`RegisteredClaims.Subject` 存 userID，含过期校验；keyfunc 断言 HMAC 方法，堵死 alg=none / 算法混淆）。
- `pkg/apperror` 新增 `Unauthorized`（UNAUTHORIZED/401）。
- `RequireAuth` 的原始解析错误只按 `request_id` 写日志，客户端统一收笼统 401（与 A3「绑定错误不外泄」同一条安全原则）。
- `login` 成功返回 `token`（`LoginResponse{id,email,token}`）。
- `internal/config/env.go`：`JWTSecret()`（`JWT_SECRET` env + dev 兜底，**secret 从不打印**）、`AllowedOrigins()`。
- `go.mod`：`golang-jwt/jwt/v5 v5.3.1` 由 `// indirect` 提升为直接依赖。

顺带（C6 同类清理）：本次涉及函数上方已完成的 `// TODO: implement` 脚手架注释一并删除（token.go/auth.go/cors.go/user.go）。`todo.go`/`repository/user.go`/`service/user.go`/`service/todo.go` 里早几天遗留的同类注释仍在，待 Day 27 工程收尾统一清。



