# Day 17 学习记录

## 日期

2026-07-10

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/17

## 今日教案

- 教案文档：`docs/daily/day-17-lesson.md`

## 核心任务

- 学习 `database/sql` 包的基础用法。
- 使用 `go-sql-driver/mysql` 驱动连接 MySQL。
- 配置数据库连接（DSN、连接池）。
- 使用 Docker Compose 启动 MySQL 开发环境。
- 实现基础 CRUD 操作（Create、Read、Update、Delete）。

## 验收标准

- Docker Compose 能启动 MySQL。
- `internal/config/database.go` 能连接 MySQL。
- `internal/repository/todo.go` 实现 CRUD。
- 至少写 3 组测试（Create、FindByID、FindAll）。
- `go test ./internal/repository` 通过（需要数据库运行）。
- 能解释 DSN、`sql.Open`、`db.Ping`、连接池配置。

## 可选挑战题

- 使用 Docker Compose 启动 MySQL。

## 答疑记录

- 待记录。

## 今日产出

- 创建 `deployments/docker-compose.yml`，启动 MySQL 8.0 开发环境。
- 创建 `deployments/migrations/001_create_todos.sql`，初始化 `todos` 表。
- 实现 `internal/config/database.go`：DSN 生成和 MySQL 连接。
- 实现 `internal/model/todo.go`：Todo 模型。
- 实现 `internal/repository/todo.go`：Create、FindByID、FindAll、Update、Delete。
- 编写数据库仓库集成测试，并通过容器内 MySQL 验证。

## 运行过的命令

```bash
docker compose -f deployments/docker-compose.yml up -d
docker exec -i go_learning_mysql mysql -uroot -ppassword go_learning < deployments/migrations/001_create_todos.sql
RUN_INTEGRATION_TESTS=true go test ./internal/repository -v
go build ./...
go test ./...
go vet ./...
```

## 代码 Review 结论

- `DatabaseConfig.DSN()` 使用标准 MySQL DSN 格式，包含 `charset=utf8mb4`、`parseTime=True` 和 `loc=Local`。
- `Connect()` 使用 `sql.Open` 创建连接池对象，并通过 `db.Ping()` 验证数据库连通性。
- `TodoRepository` 使用参数化 SQL，避免拼接字符串和 SQL 注入风险。
- `FindByID` 对 `sql.ErrNoRows` 做了特殊处理，找不到数据时返回 `nil, nil`。
- `FindAll` 使用 `rows.Next()` 遍历结果，并在最后检查 `rows.Err()`。
- `Update` 和 `Delete` 使用 `Exec` 执行写操作，并返回错误给调用方。
- 仓库集成测试通过 Docker Compose 启动的 MySQL 验证了 CRUD 行为。
- `go build ./...`、`go test ./...`、`go vet ./...` 均已通过。

## 今日小测试

1. `sql.Open` 和 `db.Ping()` 分别做什么？为什么 `Open` 不等于已经连上数据库？
   - 回答：`sql.Open` 的作用是创建一个 `*sql.DB`、初始化数据库连接池、检查驱动名称和 DSN 格式等基础配置，通常不会立即建立真实连接。`db.Ping()` 会实际尝试连接数据库，用来确认数据库服务、地址端口、用户名密码、网络连通性和驱动是否正常。
   - 结果：通过。
   - 标准答案：`sql.Open` 创建 `*sql.DB` 和连接池对象，但不保证数据库已连通；`db.Ping()` 才会真实发起连接并验证可用性。
2. DSN 中 `parseTime=True` 的作用是什么？如果去掉可能会怎样？
   - 回答：`parseTime=True` 是 MySQL Go 驱动的配置，用来把数据库中的日期时间字段 `DATETIME`、`TIMESTAMP`、`DATE` 自动解析成 Go 的 `time.Time`。
   - 结果：通过。
   - 标准答案：`parseTime=True` 让驱动把时间类型自动转换为 `time.Time`。如果去掉，扫描时间字段时可能出现类型不匹配或得到 `[]byte/string`，影响后续处理。
3. 为什么 `FindByID` 需要特殊处理 `sql.ErrNoRows`？返回 `nil, nil` 的含义是什么？
   - 回答：查询单条数据时，如果没有匹配的记录，`Scan` 会返回 `sql.ErrNoRows`。这里要把“没找到数据”和“数据库发生故障”区分开。
   - 结果：通过。
   - 标准答案：`sql.ErrNoRows` 表示查询成功但没有数据，不是数据库故障。返回 `nil, nil` 表示“没有找到记录，但没有发生错误”，让上层可以按业务逻辑处理。
4. `rows.Close()` 为什么必须用 `defer`？如果忘记关闭会有什么问题？
   - 回答：`rows` 持有数据库查询结果以及相关连接资源。`Close()` 用来释放这些资源，让连接可以回到连接池中复用。使用 `defer` 的好处是登记一次，无论后面从哪里返回，函数结束时都会自动关闭。如果忘记关闭，可能导致连接无法及时归还、连接池耗尽、后续查询等待、请求超时和连接泄漏。
   - 结果：通过。
   - 标准答案：`rows.Close()` 释放结果集和底层连接资源，必须在查询后及时关闭。`defer` 能确保所有返回路径都执行关闭，避免连接泄漏。
5. 为什么数据库操作中要使用 `?` 占位符而不是字符串拼接？
   - 回答：占位符用于安全地传递参数，防止 SQL 注入。
   - 结果：通过。
   - 标准答案：占位符由数据库驱动安全绑定参数，避免手动拼接带来的 SQL 注入风险，也减少转义和字符串拼写错误。

## 测试结果

- 5 题全部通过。

## 遇到的问题

- 本机没有 `mysql` 客户端，改用 `docker exec` 在容器内执行迁移。
- `docker-compose.yml` 中的 `mysql_data:` 是正确的命名卷声明，不需要修改。

## 关键收获

1. `database/sql` 提供统一的数据库操作接口，MySQL 驱动通过匿名导入注册。
2. DSN、连接池和 `db.Ping()` 是建立稳定数据库连接的基础。
3. CRUD 仓库层使用参数化 SQL 和 `rows.Next()` 模式，能可靠处理单行和多行查询。
4. 通过 Docker Compose 启动数据库后再跑集成测试，是验证数据库代码的实用方式。
5. `sql.ErrNoRows`、`rows.Err()`、`defer rows.Close()` 是数据库查询里最容易漏掉的细节。

## 明日计划

- 进入 Day 18：GORM 模型和 repository 实现。
