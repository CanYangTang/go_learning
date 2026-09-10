# Day 21 学习记录

## 日期

2026-09-03（教案与审计）～ 2026-09-08（实现、验证与收尾）

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/21

## 今日教案

- 教案文档：`docs/daily/day-21-lesson.md`

## 核心任务

- 端到端验证 TODO API 雏形：Docker + `make run` + `curl`，确认数据真的落库。
- 全量审计项目问题，产出 `docs/issues-backlog.md` 并按 Day 分配归属。
- 修 4 条缺陷：补 Recovery middleware（A1）、`NoRoute` 响应形状（A2）、绑定错误原文外泄（A3）、`go mod tidy`（C1）。
- 把 `docs/api/todo-api.md` 从「愿望清单」改成「已实现 + 计划中」。
- 给 `docs/architecture/todo-api.md` 补依赖方向、接口归属、依赖组装位置。
- 完成 `docs/weekly/week-03.md` 复盘（并补齐被计划遗漏的 `week-02.md`）。
- 补 README 的 MySQL 前置说明和 `curl` 示例。

## 可选挑战题

- 给架构文档画依赖方向图。
- `writeError` 兜底分支改用 `apperror.Internal`，消掉重复字面量。
- 给每个「计划中」接口标注负责的 Day。

## 验收标准

- 端到端跑通，重启服务后数据仍在。
- handler panic 时返回 500 + 统一错误信封，且日志里有一条带 request ID 的记录。
- 任意未知路径返回 `{"error":{"code":"NOT_FOUND","message":"..."}}`。
- 绑定失败的响应里不含内部 Go 类型名。
- `go mod tidy -diff` 无输出。
- 三份文档 + README 更新完成，API 文档里已实现部分的例子都能跑通。
- `make fmt`、`make test`、`make vet` 全部通过。

## 答疑记录

- 教案阶段没有提问，直接进入实现。教案本身在发布前做了自校验，过程中发现并更正了两处错误，记在「遇到的问题」里。

## 今日产出

- `internal/middleware/recovery.go` + `recovery_test.go`：`defer` + `recover()` 把 panic 转成 500 + 统一错误信封；panic 值和 `debug.Stack()` 只写日志，按 `request_id` 与访问日志关联。5 个用例覆盖信封形状、不泄漏 panic 值、日志含 request ID、Logging 仍能记到这次请求、正常请求不受影响。
- `internal/handler/notfound.go` + `notfound_test.go`：`NoRoute` 改走 `writeError`，返回 `{"error":{"code":"NOT_FOUND","message":"route not found"}}`，响应里不回显请求路径。4 个用例，其中一个显式断言旧的字符串形状 `"error":"` 不再出现。
- `pkg/apperror/error.go` 新增 `NotFound`，`error_test.go` 补上三个构造函数的 code/status 断言和 `errors.As` 可匹配性验证（`writeError` 依赖后者）。
- `internal/handler/todo.go`：绑定失败改为「原文写日志、对外返回固定文案」。`todo_test.go` 新增 `TestCreateTodoDoesNotLeakBindErrorDetails`，三个子用例（缺字段 / JSON 截断 / 传数组）断言响应里不含 `CreateTodoRequest`、`Field validation`、`cannot unmarshal`、`unexpected EOF`。
- `cmd/server/main.go`：按 `RequestID → Logging → Recovery → CORS → AuthPlaceholder` 挂载并注释了顺序理由，注册 `router.NoRoute`，删掉不再需要的 `net/http` 导入。
- `go mod tidy`：四个直接依赖不再被标成 `// indirect`。
- `docs/issues-backlog.md`：25 条全量审计（含 Day 21 新增的 C6），每条都对着代码验证过，按 P0–P3 分级并分配到具体 Day。
- `docs/api/todo-api.md` 重写成「已实现 / 计划中」两节，已实现部分的示例全部是从运行中的服务上抓下来的。
- `docs/architecture/todo-api.md` 新增「当前状态」小节：真实分层、真实 TODO 字段、中间件顺序理由，以及 Day 20 定下但从未记录的三条设计决策。
- `README.md` 新增「本地启动」：MySQL 前置步骤、`DB_DSN` 覆盖方式、五条可直接粘的 `curl`。
- `docs/weekly/week-03.md` 复盘，并补齐被计划遗漏的 `docs/weekly/week-02.md`。

## 运行过的命令

```bash
# 实现后的验证
gofmt -l .
go vet ./...
go mod tidy
go mod tidy -diff

# 端到端（手工验证必须排在集成测试之前）
docker compose -f deployments/docker-compose.yml up -d
go build -o /tmp/gl_server ./cmd/server && /tmp/gl_server
curl -s -i -H 'Origin: https://example.com' http://localhost:8080/api/v1/health
curl -s -X POST http://localhost:8080/api/v1/todos -H 'Content-Type: application/json' -d '{"title":"Day 21 end-to-end"}'
curl -s http://localhost:8080/api/v1/todos
curl -s -X POST http://localhost:8080/api/v1/todos -H 'Content-Type: application/json' -d '{}'
curl -s -X POST http://localhost:8080/api/v1/todos -H 'Content-Type: application/json' -d '{"title":'
curl -s http://localhost:8080/health
curl -s -o /dev/null -w '%{http_code}\n' -X OPTIONS http://localhost:8080/api/v1/todos \
  -H 'Origin: https://example.com' -H 'Access-Control-Request-Method: POST'
curl -s -D- -o /dev/null -H 'X-Request-ID: my-own-id-123' http://localhost:8080/api/v1/health

# 排查端口占用
lsof -nP -iTCP:8080 -sTCP:LISTEN

# 单元测试（注意：不带 RUN_INTEGRATION_TESTS 时 repository 的 8 个用例全部 t.Skip，
# ok internal/repository 不代表真实 SQL 跑过）
go test -count=1 ./...

# 真正跑数据库的那一次（teardown 会 DROP TABLE todos，所以必须排在手工验证之后）
RUN_INTEGRATION_TESTS=true go test -count=1 ./internal/repository

# 收尾
grep -rn "TODO: implement" --include="*.go" .
```

## 代码 Review 结论

- `Recovery` 抓住了最关键的一点：panic 值可能带主机名、路径、SQL 片段，它同时又是唯一的调试线索 —— 所以进日志、不进响应，并用 request ID 把 panic 日志和访问日志串起来。
- `NotFoundHandler` 放在 `internal/handler` 包内是对包边界的正确处理：`writeError` 是 unexported 的，只有同包才能复用这个唯一错误出口。`internal/middleware/recovery.go` 拿不到它，改用 `pkg/response` 重建同形状信封，并在注释里写明了原因，没有为了复用去导出 `writeError`。
- 绑定失败的处理是「日志详细、响应固定」的标准形态，日志里能看到 `bind_error=unexpected EOF`，客户端只拿到固定文案。
- 中间件挂载顺序带了注释说明理由，这比顺序本身更有价值 —— 下次有人调顺序时会先看到代价。
- `main.go` 的装配链保持单一位置，删掉了失效的 `net/http` 导入。
- 唯一的返工：`pkg/apperror/error.go` 的 `// TODO: implement` 脚手架注释没清（详见「遇到的问题」）。
- 两点后续（不影响今天验收）：
  - `Recovery` 兜不住「handler 已写了部分响应体再 panic」的情况，此时 `AbortWithStatusJSON` 会二次写 header，Gin 会打 `Headers were already written`，客户端拿到半截 JSON 加一段 500 body。当前没有这种路径。
  - `"invalid request body: title is required"` 这个文案略微过于具体：JSON 截断和传数组两种情况其实与 `title` 无关，等 Day 22 加上新字段后这句话会变成错的。测试只断言「不泄漏」不断言措辞，改文案不会挂测试。

## 今日小测试

1. `router.Use(RequestID(), Logging(), Recovery(), CORS(), AuthPlaceholder())` —— 如果把 `Recovery()` 挪到 `Logging()` 前面，一次 panic 请求的行为会有什么变化？客户端看到的和日志里看到的分别是什么？为什么？

- 回答：一次 500 只在 panic 日志里出现、访问日志里查不到，排查时会显得像「请求根本没到服务」。这就是 Recovery 要在 Logging 之内的唯一理由，和「谁更重要」无关。
- 结果：通过。
- 标准答案：客户端看到的**完全不变** —— 两种顺序都是 500 + `{"error":{"code":"INTERNAL_ERROR",...}}`，响应字节一致。变的只有日志：`Recovery` 在外层时只有 1 行 `panic request_id=... value=...`，`Logging` 的 `c.Next()` 后续代码（读 `c.Writer.Status()`、算 latency、打访问日志）在 panic 解栈时已经被掀掉了，永远执行不到；`Logging` 在外层时是 2 行，panic 日志 + `request_id=... method=GET path=/boom status=500 latency=0s`。作答抓住的正是要点：**这个差异从外部观测不到**，所以它不会在联调时暴露，只会在事后排查时表现为「这个请求在日志里不存在」。

2. `Recovery` 里用的是 `c.AbortWithStatusJSON`。如果改成 `c.JSON`（不 Abort），这次 panic 请求的响应和后续 handler 的执行会不一样吗？`Abort` 在这里到底起了什么作用？

- 回答：panic 在最后一个 handler（我们线上就是这种：业务 handler 挂了）时两者无差别，响应字节完全相同。panic 在 Recovery 之内、但不是最后一环的中间件（CORS / AuthPlaceholder）时差别是实打实的：不 Abort 的话，后面的中间件和业务 handler 照常执行。
- 结果：通过，**并且纠正了教案**。教案第一版说「不 Abort 后续 handler 会继续跑」，我实测了「业务 handler 内 panic」这一种场景后改成第二版「`Abort` 只是惯例」—— 这个更正本身是错的，被这份作答推翻。已按作答重写教案「改动二」。
- 标准答案：取决于 **panic 发生在哪一环**，三种位置三种行为。机制在 `c.Next()`：它是 `c.index++` 然后 `for c.index < len(handlers) { handlers[c.index](c); c.index++ }`，panic 解栈后控制流回到**外层**的那个 `Next()` 循环，循环继续 `index++`。实测（临时包，跑完即删）：

  ```text
  panicBeforeNext  abort=true  status=500 后续中间件=false 路由handler=false body={"code":"INTERNAL_ERROR"}
  panicBeforeNext  abort=false status=500 后续中间件=true  路由handler=true  body={"code":"INTERNAL_ERROR"}route ok
  panicAfterNext   abort=true  status=200 后续中间件=true  路由handler=true  body=route ok{"code":"INTERNAL_ERROR"}
  panicAfterNext   abort=false status=200 后续中间件=true  路由handler=true  body=route ok{"code":"INTERNAL_ERROR"}
  ```

  - panic 在某个中间件**调用 `c.Next()` 之前**：`c.index` 还停在那个中间件上，恢复后 `index++` 正好落在下一个 handler，后续中间件和业务 handler 照常执行，响应体拼成 `{"code":"INTERNAL_ERROR"}route ok` 这种坏 JSON。`Abort` 把 `index` 设成 63 才拦住 —— **这里 `Abort` 是必需的，不是惯例。**
  - panic 在业务 handler（链的最后一环）：`index` 已到末尾，加不加 `Abort` 都一样。这是作答里说的「我们线上就是这种」，也是我第一次只测了这一种。
  - panic 在某个中间件 `c.Next()` **之后**的收尾代码里：业务 handler 已经写完 200 和 body，`Recovery` 只能追加，**状态码停在 200**，gin 打 `Headers were already written`。这种 `Recovery` 兜不住，`Abort` 也救不了。

  方法论上的教训：**测一种场景就下结论，等于把「我想到的那种情况」当成了全部。**

3. `internal/handler/notfound.go` 放在 `internal/handler` 包里，而 `internal/middleware/recovery.go` 明明也要返回同样形状的错误信封，却自己用 `pkg/response` 重建了一遍。为什么不把 `writeError` 导出成 `WriteError` 让两边都用？

- 回答：硬约束：会形成 import 环。今天 `internal/handler/todo.go` 已经 import `internal/middleware`（用 `middleware.RequestIDFromContext` 记 bind 错误日志），如果 middleware 反过来 import handler 拿 `WriteError`，Go 直接编译不过。
- 结果：通过，而且比我在 review 和架构文档里的说法更准确 —— 我写的是「包边界」，作答给出的是编译期的硬失败。
- 标准答案：直接原因就是作答说的 **import 环**：`handler → middleware`（`RequestIDFromContext`）已经存在，再加 `middleware → handler` 就是环，Go 拒绝编译，这是硬约束不是风格问题。补一层：即便没有环，`middleware → handler` 也是**错的依赖方向**。middleware 是比 handler 更底层的横切设施，它不应该知道 handler 包的存在；真要共享错误出口，正确做法是把它下沉到两边都能依赖的地方（`pkg/response` 或一个新的 `pkg/httperr`），而不是让下层反向依赖上层。当前选择是「middleware 用 `pkg/response` 重建同形状信封 + 注释写明原因」，代价是信封形状有两处定义、改动时要同步 —— 这个代价被记在了注释里，是有意接受的。

4. 修改之前，`POST /api/v1/todos` 传 `{}` 会把什么内容返回给客户端？这里面哪一部分算是泄漏、为什么算？改完之后，如果一个客户端报告说「我这个请求返回 400 但我不知道为什么」，你还能怎么查？

- 回答：`{"error":{"code":"VALIDATION_ERROR","message":"Key: 'CreateTodoRequest.Title' Error:Field validation for 'Title' failed on the 'required' tag"}}`。算泄漏的部分：`CreateTodoRequest` —— 内部 Go 结构体名，它不是 API 契约的一部分，改名重构就会改变对外响应，等于把内部命名焊死在了 API 上。`Field validation for ... on the 'required' tag` —— go-playground/validator 的固定措辞，等于向未认证的调用方公示了技术栈和校验实现。传数组时那条 `json: cannot unmarshal array into Go value of type handler.CreateTodoRequest` 还额外暴露了包名 `handler`。为什么算泄漏而不只是「难看」：这些信息对合法客户端零价值（它拿不到可编程处理的字段级错误，只有一句英文散文），对攻击者却是侦察素材 —— 技术栈指纹、结构体命名风格、可以据此猜其他字段名。信息不对称是反的。改完之后怎么查：让对方回报响应头里的 `X-Request-ID`（`RequestID` 中间件对每个请求都写，`request_id.go:25`；客户端自己带这个头的话会沿用），然后在服务端日志里搜这个 ID，会有两条：`request_id=<id> bind_error=json: cannot unmarshal array into Go value of type handler.CreateTodoRequest` / `request_id=<id> method=POST path=/api/v1/todos status=400 latency=0s`。第一条是原始的 validator/json 错误（原文一字不少，只是留在了服务端），第二条给出方法、路径、状态码、耗时。这就是「详细的进日志、按 `request_id` 可回溯，对外只给固定文案 + 错误码」这套做法的完整闭环 —— 诊断能力没有丢失，只是换了个投递地址。
- 结果：通过，完整。「信息不对称是反的」这句是这条的核心：判断一段错误文本该不该外露，看的是它对合法客户端和对攻击者分别值多少，而不是看它是否敏感。
- 标准答案：与作答一致，无需补充。三种坏输入的原文分别是 validator 的 `Key: 'CreateTodoRequest.Title' ...`、JSON 截断的 `unexpected EOF`、类型不符的 `json: cannot unmarshal array into Go value of type handler.CreateTodoRequest`；泄漏的是内部结构体名、包名和校验库指纹。追查链条就是作答写的那两条日志 —— 唯一值得强调的是**它依赖 `RequestID` 中间件在最外层**，如果 `RequestID` 挂在 `Logging` 之内，bind 错误日志里的 request ID 会是兜底值 `request-id`，两条日志就串不起来，这条追查路径当场失效。

5. 今天的手工 `curl` 验证必须排在 `go test -count=1 ./...` **之前**，原因是什么？如果顺序反过来，你会在哪一步得到什么现象，并可能得出什么错误结论？

- 回答：因为集成测试会拆掉手工验证依赖的那张表。`internal/repository/todo_test.go` 的 `TestMain` teardown 执行 `DROP TABLE IF EXISTS todos`（第 38 行），每个用例开头还会 `DELETE FROM todos`（第 49 行）。而它用的 DSN（第 22 行硬编码）和 `config.DefaultDSN()` 指向同一个 `go_learning` 库 —— 测试和开发服务共用一个数据库，没有隔离。
- 结果：部分通过。机制、三个行号（`:22`、`:38`、`:49`）和「共用一个库、没有隔离」都对，但**漏了 `RUN_INTEGRATION_TESTS` 门控**，也没说出反过来会看到什么现象、会得出什么错误结论。
- 标准答案：分两层。
  1. **门控（漏掉的那一层）**：`TestMain`（`:18`）看不到 `RUN_INTEGRATION_TESTS=true` 就直接 `os.Exit(m.Run())`，`testDB` 保持 nil，8 个用例全部 `t.Skip`。所以**裸 `go test -count=1 ./...` 既不建表也不删表**，题目里那个「必须排在前面」的顺序要求，只对 `RUN_INTEGRATION_TESTS=true go test ./internal/repository` 成立。这一层比顺序本身更重要：它意味着 `ok internal/repository` 这行输出**一行 SQL 都没跑过**，「测试全绿」这个信号本身是失真的（backlog C5）。我自己也在这条上错过 —— 先前汇报测试结果时说过「含 repository 的真实 SQL」。
  2. **顺序（答对的那一层）**：一旦带上门控变量，teardown 的 `DROP TABLE IF EXISTS todos` 会把手工 POST 进去的数据连表一起删掉。反过来跑的现象是：集成测试 `ok`，然后 `curl http://localhost:8080/api/v1/todos` 返回 **500 + `{"error":{"code":"INTERNAL_ERROR",...}}`**（GORM 报 `Error 1146: Table 'go_learning.todos' doesn't exist`，被 service 包成 `apperror.Internal`），而不是空数组。错误结论就是「持久化坏了 / 分层写错了 / GORM 配置有问题」，而真实原因是表被自己的测试删了。更隐蔽的变体是重启服务后再 curl：`AutoMigrate` 会把表重新建好，于是 `GET` 返回 `[]`，看起来像「数据写进去了但读不出来」或「重启丢数据」—— 这正好是端到端验证要证伪的那个结论，会被误判成持久化失效。

## 测试结果

- `gofmt -l .`：无输出。
- `go vet ./...`：无输出。
- `go mod tidy -diff`：无输出，四个直接依赖不再是 `// indirect`。
- `go test -count=1 ./...`：全部 `ok`（repository 的 8 个用例因 `RUN_INTEGRATION_TESTS` 未设置全部 `t.Skip`，不代表真实 SQL 跑过）。
- `RUN_INTEGRATION_TESTS=true go test -count=1 ./internal/repository`：7 个用例全部 `PASS`，`ok ... 0.884s`，真实连了 MySQL。
- 端到端 `curl`：健康检查、`POST /todos`（201，非 0 id）、`GET /todos`、重启服务后数据仍在、`{}` 和截断 JSON 返回固定文案 400、未知路径返回 `NOT_FOUND` 404、`OPTIONS` 预检 204、`X-Request-ID` 透传，八项全部实测通过（详见「运行过的命令」）。
- 小测：Q1/Q3/Q4 通过，Q2 纠正了教案（已按作答改写），Q5 部分通过（漏了 `RUN_INTEGRATION_TESTS` 门控）。

## 遇到的问题

- **教案里发布了错误的中间件顺序。** 原本写的是 `RequestID → Recovery → Logging`，理由是「Recovery 必须最外层」。写测试时才想清楚：Recovery 在外层意味着 panic 在到达 `Logging` 的 `c.Next()` 后续代码之前就被截住了。用临时包实测（跑完即删）：Recovery 在外 → 只有 1 行 panic 日志；Logging 在外 → 2 行，panic 日志 + `status=500 latency=0s` 的访问日志。已改成 `RequestID → Logging → Recovery → CORS → AuthPlaceholder`。
- **教案里还发布了一个错误的断言，而且更正了一次还是错的。** 第一版说 `c.JSON` 不加 `Abort` 后面的 handler 还会执行；实测一种场景（业务 handler 里 panic）发现两种写法响应一致，于是改成第二版「`Abort` 只是惯例，不是修 bug」。**第二版被小测第 2 题的作答纠正了**：panic 发生在某个中间件调用 `c.Next()` **之前**时，不 `Abort` 后续中间件和业务 handler 真的会照常执行，响应体拼成 `{"code":"INTERNAL_ERROR"}route ok` 这种坏 JSON。根因是我只测了一种 panic 位置就把结论推广了 —— panic 位置有三种（中间件 Next 前 / 业务 handler 内 / 中间件 Next 后），三种行为都不同，我两次都只测了一种。实测数据见教案「改动二」。
- **`ok internal/repository` 一直是假绿。** `TestMain`（`todo_test.go:18`）没有 `RUN_INTEGRATION_TESTS=true` 就直接 `m.Run()`，`testDB` 保持 nil，8 个用例全部 `t.Skip`。我先前汇报 `go test -count=1 ./...` 的结果时说「含 repository 的真实 SQL」，这是错的 —— 那次一行 SQL 都没跑。补跑 `RUN_INTEGRATION_TESTS=true go test -count=1 ./internal/repository`，7 个用例才真正过。这条比 A1/A2/A3 更值得警惕：它让「测试全绿」这个信号本身失真，已补进 backlog C5。
- **`pkg/apperror/error.go` 的 `// TODO: implement` 注释填完实现后没删。** 这是 Day 20 记过的同一个问题第三次出现（Day 18 的 `internal/repository/todo.go`、Day 20、今天）。顺手 `grep` 了一遍，发现 `week02-core` 里还躺着 8 处，已作为 C6 记入 backlog。以后填完实现 `grep -rn "TODO: implement"` 应该是收尾动作的固定一步。
- **`go mod tidy` 漏跑了。** 说了「这条你自己跑一下」，结果两边都没跑，`go mod tidy -diff` 一查还在。
- **端口上挂着旧进程，第一轮端到端验证全是假的。** `:8080` 被 9 月 3 日 `make run` 留下的进程占着（`ps` 显示是 go-build 缓存里的旧二进制），新编的二进制 `bind: address already in use` 直接退出 —— 但我只 `sleep 2` 就开始 curl，没看启动日志。结果 A3 的响应还是泄漏 validator 原文，看上去像完全没修。`lsof -nP -iTCP:8080` 才发现真相。教训：验证前先确认端口上跑的是哪个二进制，或者干脆先看服务启动日志。

## 关键收获

1. **中间件的「内外」是由 panic 怎么解栈决定的，不是由「谁更重要」决定的。** panic 会一路掀掉调用栈，第一个 `defer` + `recover()` 命中的地方就是解栈的终点，比它更外层的中间件的 `c.Next()` 后续代码全都不会执行。所以 `Recovery` 要在 `Logging` **之内**才能保住访问日志，要在 `CORS` **之外**才能让 500 保留跨域头。顺序不是风格问题，是行为问题。
2. **面向客户端的错误信息和面向自己的错误信息是两回事。** validator 原文、gorm 错误文本、panic 值都属于后者 —— 它们会带内部类型名、表名、SQL 片段、主机名，对合法客户端毫无用处，对攻击者是地图。做法固定：详细的进日志、按 `request_id` 可回溯；对外只给固定文案 + 错误码。
3. **绿色的测试不等于系统能跑。** Week 3 的检查点写了「TODO API 雏形可以运行」，但所有单元测试都用替身，一行 SQL 都没经过，端到端直到今天才第一次真跑。而真跑一次立刻暴露了三件事：缺 Recovery、404 形状不一致、绑定错误外泄 —— 三条都是单元测试全绿的情况下存在的。
4. **文档漂移比没文档更危险。** 一周就漂到 API 文档描述 7 个不存在的接口、架构文档还写着「初期可使用内存实现」。判断标准可以很机械：文档里每个 `curl` 现在跑一遍，得不到文档写的结果就说明漂了；做不到的移进「计划中」并标上负责的 Day。
5. **把问题写下来并分配 Day，比当场修完更重要。** 25 条里今天只修了 8 条，其余按 P0–P3 分级挂到了具体的 Day 上。P0/P1 当天修，P2 挂在「下一步开发之前」（比如 CORS 白名单必须排在 JWT 之前），P3 攒到工程收尾。清单本身成了 Day 22–28 的输入。

## 明日计划

- Day 22：数据库设计。开工前先做完两个决策 —— `Done bool` 还是 `Status string`（backlog D3），以及 schema 以 `AutoMigrate` 为准还是以迁移脚本为准（C2）。
- 顺带处理 C5：集成测试的硬编码 DSN 和 `DROP TABLE` teardown。
