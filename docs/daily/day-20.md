# Day 20 学习记录

## 日期

2026-09-01

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/20

## 今日教案

- 教案文档：`docs/daily/day-20-lesson.md`

## 核心任务

- 重构为 handler / service / repository 三层。
- 新增 `internal/service/`，把业务规则和错误包装收口到 service。
- handler 通过接口依赖 service，不再返回假数据。
- 在 `cmd/server/main.go` 完成依赖组装。
- 可选挑战：为 service 增加不依赖数据库的单元测试。

## 验收标准

- `internal/service/` 里有 `TodoService`，包含业务校验和错误包装。
- `TodoRepository` 接口定义在 service 包，不在 repository 包。
- handler 通过接口依赖 service。
- `cmd/server/main.go` 完成 repo → service → handler 的组装。
- service 层单元测试不依赖数据库。
- `make fmt`、`make test`、`make vet` 全部通过。

## 答疑记录

- 教案阶段没有提问，直接进入实现。
- 实现后确认：`make fmt test vet` 不需要启动 MySQL，所有单元测试都用替身；数据库只在 `make run` 做端到端验证时才需要。

## 今日产出

- 新增 `internal/service/todo.go`，`TodoService` 承担 title 校验、错误包装和空切片归一化。
- 新增 `internal/service/todo_test.go`，7 个用例，用内存假 repository，不依赖数据库。
- 新增 `internal/config/gorm.go`，提供 `DefaultDSN()` 和 `ConnectGorm()`，repository 需要的是 `*gorm.DB`。
- 新增 `internal/config/gorm_test.go`，校验 DSN 拼装。
- 改造 `internal/handler/todo.go`，通过接口依赖 service，新增响应转换 `newTodoResponse` 和统一错误出口 `writeError`。
- 重写 `internal/handler/todo_test.go`，9 个用例，用假 service 覆盖成功、绑定失败和两类业务错误。
- 更新 `cmd/server/main.go`，完成 repo → service → handler 组装，注册 `/api/v1/todos` 两条路由。
- 清理 `internal/repository/todo.go` 和 `cmd/server/main.go` 里 Day 18、Day 20 遗留的脚手架注释。

## 运行过的命令

```bash
make fmt
make vet
make test
go test -count=1 ./internal/...
```

## 代码 Review 结论

- 分层边界正确：service 里没有 `gin` 也没有 `gorm`，handler 里没有 `gorm`。
- `TodoRepository` 接口定义在 service 包，只声明用得到的 2 个方法，`*repository.TodoRepository` 隐式满足它。
- `CreateTodo` 校验失败时直接返回，不会白跑一次数据库，测试里用 `createCalls` 断言了这一点。
- 底层错误统一包成 `apperror`，handler 靠 `errors.As` + `appErr.StatusCode` 映射状态码，没有 if-else 判断错误类型。
- `make([]Todo, 0, len(todos))` 保证空列表序列化成 `[]` 而不是 `null`。
- 依赖只在 `main` 组装，三层内部都不自己 new 依赖。
- 已修正：`internal/service/todo.go` 的 import 分组（标准库和项目包应分开，`gofmt` 只排序不分组）。
- 已清理：实现完成后残留的 `// TODO: implement` 脚手架注释。

## 今日小测试

1. 为什么 `TodoRepository` 接口要定义在 `internal/service/` 而不是 `internal/repository/`？
   - 回答：因为接口应该由「使用它的一方」定义。
   - 结果：正确。
   - 标准答案：Go 的惯例是调用方定义接口。如果接口定义在 repository 包，service 就必须 import repository 才能拿到接口，依赖又绑回了具体实现；定义在 service 包后，service 只依赖自己声明的抽象，测试时能随手换成假实现，也能只声明用得到的方法。
2. `repository.TodoRepository` 有 6 个方法，service 里的接口只声明了 2 个。这样做的好处是什么？为什么不会编译报错？
   - 回答：这是「最小接口」原则，降低依赖范围；不报错是因为 Go 接口是隐式实现的。
   - 结果：正确。
   - 标准答案：补充两点收益。一是测试里的假 repository 只需要实现 2 个方法，不用为用不到的方法写空壳；二是 repository 以后增删其他方法不会波及 service。Go 不需要 `implements` 声明，方法集覆盖了接口就自动满足。
3. 如果 service 直接把 `gorm` 的错误 return 出去，handler 那边会出现什么后果？
   - 回答：识别不出具体的业务错误信息。
   - 结果：方向对，漏了两个具体后果。
   - 标准答案：handler 的 `errors.As(err, &appErr)` 匹配不上，只能落到兜底分支，于是参数错误、资源不存在全都变成 500，状态码语义丢失。另一个风险是如果把原始错误信息写进响应，`gorm` 的错误文本（可能含表名、SQL 片段）会泄漏给客户端。
4. `ListTodos` 里那句 `if todos == nil { todos = []model.Todo{} }` 不写会怎样？具体影响谁？
   - 回答：如果查询没有任何数据，repository 可能返回 nil slice，直接序列化结果是 null。影响 handler 的 JSON 响应、前端调用方、直接调用 service 的其他代码，以及 service 层自身「列表一定可遍历」的约定。
   - 结果：正确，回答很完整。
   - 标准答案：同上。教案里实测过：`[]model.Todo{}` 序列化成 `"data":[]`，nil slice 序列化成 `"data":null`。本项目 handler 里的 `make([]Todo, 0, len(todos))` 其实也能兜住这一层，但归一化放在 service 更靠前，保护的是所有调用方而不只是 HTTP 那一个。
5. `NewTodoService`、`NewTodoHandler` 都在 main 里调用，为什么不让 handler 自己 new 一个 service？
   - 回答：因为 handler 不应该负责组装底层依赖。main 组装、handler 管 HTTP、service 管业务、repository 管数据库。
   - 结果：正确。
   - 标准答案：补充一个更硬的理由：handler 自己 new 就意味着它依赖具体类型而不是接口，测试时无法注入假 service，只能连着真业务甚至真数据库一起测。依赖统一在入口组装，才能保证每层都能被单独替换和单独测试。

得分：5 题里 4 题完全正确，第 3 题方向正确但不够具体。

## 测试结果

- `make fmt` 通过（顺带补齐了 Day 17、Day 18 几个文件缺失的行尾换行）。
- `make vet` 通过。
- `make test` 通过，`go test -count=1 ./internal/...` 全部包 ok。
- service 和 handler 的单元测试不依赖数据库；repository 集成测试仍由 `RUN_INTEGRATION_TESTS` 门控。

## 遇到的问题

- `internal/service/todo.go` 的 import 把 `strings` 排在了项目包之后。`gofmt` 只在组内排序，不会重新分组，需要 `goimports` 或编辑器保存时整理。
- 实现完成后脚手架的 `// TODO: implement` 注释没有清理，会让人误以为代码未完成。`internal/repository/todo.go` 从 Day 18 起就有同样的遗留，本次一并清掉。
- 本机 Docker（colima）未启动，端到端 curl 验证未做。今天的验收标准是 `make fmt test vet`，不受影响。

## 关键收获

1. 引入 service 层的真正理由是「业务规则无处安放」，不是为了目录看起来规范。
2. Go 的接口归属和 Java 相反：接口定义在调用方，而且只声明用得到的方法。
3. 错误要在 service 层就包成带状态码的 `apperror`，handler 才能不做类型判断地统一映射。
4. 分层最直接的回报是可测性：service 用假 repository、handler 用假 service，两层都不需要数据库。
5. 依赖只在 `main` 组装这一条纪律，是上面所有替换和测试能成立的前提。

## 明日计划

- 进入 Day 21：周总结，完成 API 服务雏形和 Week 3 复盘。
