# Day 18 学习记录

## 日期

2026-08-13

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/18

## 今日教案

- 教案文档：`docs/daily/day-18-lesson.md`

## 核心任务

- 学习 GORM 的基本概念和使用方式。
- 使用 GORM 定义 TODO 模型和表映射。
- 使用 GORM 实现 TODO repository 的 CRUD。
- 理解 `AutoMigrate`、`gorm.ErrRecordNotFound` 和事务。

## 验收标准

- `internal/model/todo.go` 使用 GORM 定义 Todo 模型。
- `internal/repository/todo.go` 使用 GORM 实现 CRUD。
- 至少写 3 组 repository 测试。
- `go test ./internal/repository` 通过。
- `make test` 通过。
- 能解释 GORM 模型、`AutoMigrate`、`ErrRecordNotFound`、事务。

## 可选挑战题

- 增加一个事务示例：`CreateAndMarkDone`。

## 答疑记录

- 待记录。

## 今日产出

- 将 `internal/model/todo.go` 改为 GORM 模型，使用 `gorm:"primaryKey"`、`gorm:"size:255;not null"` 等标签。
- 将 `internal/repository/todo.go` 改为 GORM repository，提供 Create、FindByID、FindAll、Update、Delete 和事务方法骨架。
- 将 `internal/repository/todo_test.go` 改为 GORM 集成测试，覆盖 CRUD 和事务场景。
- 使用 Docker Compose 启动 MySQL，并通过 GORM 自动迁移验证数据库表结构。

## 运行过的命令

```bash
docker compose -f deployments/docker-compose.yml up -d
RUN_INTEGRATION_TESTS=true go test ./internal/repository -v
go build ./...
go test ./...
go vet ./...
```

## 代码 Review 结论

- `Todo` 模型使用 GORM tag 定义主键、字段长度、非空约束和软删除字段，结构清晰。
- `TodoRepository` 通过 `*gorm.DB` 完成 CRUD，`Create` 会自动回填 ID。
- `FindByID` 使用 `gorm.ErrRecordNotFound` 处理“未找到记录”的正常情况。
- `FindAll`、`Update`、`Delete` 的实现都直接使用 GORM 的常规 API，逻辑简洁。
- 事务示例 `CreateAndMarkDone` 使用 `db.Transaction` 包裹多步写操作，便于后续扩展。
- 集成测试通过 Docker Compose 启动的 MySQL 和 `AutoMigrate` 验证了 repository 行为。
- `go build ./...`、`go test ./...`、`go vet ./...` 均已通过。

## 今日小测试

1. GORM 模型里的 `gorm:"primaryKey"`、`gorm:"size:255;not null"`、`gorm:"index"` 分别表示什么？
   - 回答：`primaryKey` 把该字段设为表的主键；`size:255;not null` 表示字段长度为 255 且不能为空；`index` 给该字段创建数据库索引，加快按该字段查询/过滤的速度。
   - 结果：通过。
   - 标准答案：`primaryKey` 标记主键；`size:255` 设置列长度；`not null` 设置非空约束；`index` 创建索引。
2. `AutoMigrate` 会做什么，不会做什么？为什么不能把它当成完整迁移工具？
   - 回答：会创建不存在的表、添加新增字段、添加缺失索引和有限支持外键；不会删除已移除字段、不会可靠修改字段类型、不会重命名列、不处理数据迁移、不生成迁移历史且无法回滚。真实项目 schema 变更需要版本化、可回滚、可审查，而 AutoMigrate 只是追加式同步。
   - 结果：通过。
   - 标准答案：`AutoMigrate` 适合开发期自动创建表和补字段，但不会安全处理删除、重命名、复杂类型变更和数据迁移，也没有版本记录和回滚能力，因此不能替代正式迁移工具。
3. `gorm.ErrRecordNotFound` 和普通数据库错误有什么区别？
   - 回答：`gorm.ErrRecordNotFound` 是 GORM 定义的特殊哨兵错误，表示 SQL 执行成功了，只是没有匹配记录。它和数据库连接失败、SQL 语法错误、字段类型不匹配这类真正错误性质不同。
   - 结果：通过。
   - 标准答案：`gorm.ErrRecordNotFound` 表示查询没有命中记录，是可预期的业务分支；普通数据库错误通常表示连接、语法、权限、类型等故障，需要按错误处理。
4. `db.Transaction(func(tx *gorm.DB) error { ... })` 的回调里为什么要 `return err`？
   - 回答：GORM 靠回调返回值判断事务提交还是回滚。如果吞掉错误只打印不 return，GORM 不知道出错，会 Commit，导致部分失败的操作被错误提交。所以 `return err` 是触发回滚的唯一信号。
   - 结果：通过。
   - 标准答案：`Transaction` 回调返回 `nil` 时提交，返回非 nil error 时回滚。必须把内部错误 return 出去，否则事务会被当成成功提交。
5. `Create` 方法为什么能自动回填 `todo.ID`？
   - 回答：`Create` 接收 `*model.Todo` 指针，插入后数据库自增主键生成新 id，GORM 拿到自增值后通过指针写回 `todo.ID`。这依赖 ID 字段是主键，以及传入的是指针。
   - 结果：通过。
   - 标准答案：GORM 知道 `ID` 是主键字段，插入后读取数据库生成的自增 ID，并通过传入的结构体指针写回调用方的 `todo.ID`。

## 测试结果

- 5 题全部通过。

## 遇到的问题

- `gorm.io/driver/mysql` 只需要在数据库连接层使用，不应该放在 repository 文件里，否则会出现未使用导入。
- GORM 查询找不到记录时会打印 `record not found` 日志，这属于预期行为，需要在 repository 层单独处理。

## 关键收获

1. GORM 把 struct 和数据库表映射起来，减少手写 SQL 量，但仍需要理解模型标签和查询语义。
2. `AutoMigrate` 适合开发期自动建表和补字段，不适合替代正式数据库迁移方案。
3. `gorm.ErrRecordNotFound` 需要和真实错误区分处理，避免把“没数据”当成故障。
4. 事务可以把多步写操作包装成一个原子单元，失败时自动回滚。
5. 仍然要注意 repository 的测试数据隔离和清理，避免测试相互污染。

## 明日计划

- 进入 Day 19：日志、认证、CORS 中间件。
