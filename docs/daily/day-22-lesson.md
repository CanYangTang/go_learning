# Day 22 教案：数据库设计定案

## 学习目标

学完今天，需要能够做到：

1. 说清 `Done bool` 和 `Status string` 两条路线各自的代价，并对着现有代码给出一个可以论证的选择，而不是凭直觉挑一个。
2. 说清 `AutoMigrate` 和迁移脚本作为 schema 权威来源时分别意味着什么，选一个并说明为什么另一个会被降级。
3. 设计 `User` 表结构，并说清它和 `Todo` 之间的外键关系将怎么落地（哪一天、哪一层加）。
4. 把这两条决策和 User 设计写进 `docs/architecture/todo-api.md`，替换掉当前「待决策」的占位段落。

今天没有新代码，产出是决策 + 文档。这是 Issue #22 明确写的范围：deliverable 只有 `docs/architecture/todo-api.md`。

---

## Day 22 的位置

Week 3 收尾时，`docs/issues-backlog.md` 里留了两条明确挂到今天的决策债：

- **D3**：架构文档规划的 TODO 状态是 `Status string`（pending/done），代码已经落成了 `Done bool`。这不是漂移，是一个从 Day 1 就没做的决定。
- **C2**：`deployments/migrations/001_create_todos.sql` 缺 `deleted_at`，而且从来没有被执行过——实际建表一直靠 `cmd/server/main.go` 里的 `db.AutoMigrate(&model.Todo{})`。项目里有两套 schema 来源，一套是死的。

这两条不定下来，Day 23 写 `User` 表、Day 24 补 CRUD 时都要在两套假设之间反复横跳。今天把它们定死。

`docs/issues-backlog.md` 里还有一条 **C5**（集成测试硬编码 DSN、teardown 会 `DROP TABLE`）也排在 Day 22 附近，但那是 Day 24 补 CRUD 时顺手处理的范围——今天只做设计决策和文档，不碰测试代码。

---

## 第一件事：`Done bool` 还是 `Status string`

### 现状

`internal/model/todo.go`：

```go
type Todo struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Title     string         `gorm:"size:255;not null" json:"title"`
	Done      bool           `gorm:"default:false" json:"done"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

`docs/architecture/todo-api.md` 里「核心模型规划」写的是 `Status string`（pending/done），从 Day 1 定下来就没改过，代码则一直是 `Done bool`（Day 18 引入）。两边独立生长了三周，没人回头核对过。

### 两条路线的代价

**选 `Status string`：**

- 需要一次数据迁移：`ALTER TABLE todos ADD COLUMN status VARCHAR(20)`，把现有的 `done bool` 值映射成 `pending`/`done` 两个字符串，再删掉 `done` 列（或者保留一段时间做双写过渡）。
- `internal/service/todo.go`、`internal/handler/todo.go`、`docs/api/todo-api.md` 里所有 `"done": false` 的字段和示例都要改。
- 换来的是「多状态」的余地——将来加 `archived`、`in_progress` 只需要加一个字符串值，不需要改列类型。
- 但**当前没有任何需求需要第三种状态**。TODO 只有「完成」和「未完成」两种真实语义，`docs/api/todo-api.md` 规划的接口里也从没出现过第三种状态的使用场景。

**保留 `Done bool`：**

- 零迁移成本，`AutoMigrate` 已经建好了这个字段，所有已写的代码（service 的 `TrimSpace` 校验、handler 的 JSON 序列化、`docs/api/todo-api.md` 的已实现示例）全部不用动。
- 语义上更精确：一个只有两种取值的字段用 `bool` 比用只允许两个合法值的 `string` 更不容易出错——`bool` 不存在 "Done"/"done"/"DONE" 这种拼写不一致的问题，编译器会拦住任何非法赋值，`string` 不会。
- 代价是如果将来真的需要第三种状态，这次省下的迁移成本要连本带利还回去。

### 决策：保留 `Done bool`

理由是 YAGNI（You Aren't Gonna Need It）：架构文档规划 `Status string` 时没有配套写过任何需要第三种状态的场景，三周的实际使用（Day 18 至今）也没有产生这种需求。为一个假设的未来需求现在就承担迁移成本和字段类型退化（更宽松的 `string` 取代更严格的 `bool`），不划算。

如果将来真的出现第三种状态的需求，那时候再迁移——**且那次迁移的输入更明确**：你会知道具体要加哪个状态、它和现有两个状态如何互斥或共存，而不是像今天这样凭空猜一个 `pending`/`done` 的二元占位。

这条决策关闭 `docs/issues-backlog.md` D3。

---

## 第二件事：`AutoMigrate` 还是迁移脚本

### 现状

两套 schema 来源并存：

```sql
-- deployments/migrations/001_create_todos.sql
CREATE TABLE IF NOT EXISTS todos (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    done BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

这份 SQL **缺 `deleted_at`**（`model.Todo` 用了 `gorm.DeletedAt` 做软删除），而且——这是更根本的问题——**它从来没有被执行过**。真正建表的是 `cmd/server/main.go` 里的：

```go
db.AutoMigrate(&model.Todo{})
```

`RUN_INTEGRATION_TESTS=true` 模式下的 `internal/repository/todo_test.go` 也是调 `testDB.AutoMigrate(&model.Todo{})` 建表，同样没碰过那份 SQL 文件。

### 两条路线的代价

**切到正式迁移工具（比如 golang-migrate）：**

- 需要引入新依赖、写版本化的迁移文件、在 `make run` 或 CI 里插入迁移步骤，`AutoMigrate` 调用要删掉。
- 换来版本记录、可回滚、支持删除列/改类型这些 `AutoMigrate` 做不到的操作。
- 这些能力在生产系统里必要，但这是一个学习项目，`model.Todo` 至今只加过字段、没删过列，`AutoMigrate` 的限制从没被真实触发过。

**继续用 `AutoMigrate`：**

- 零迁移成本，`main.go` 现在的写法就是最终写法。
- 代价是明确的：不能删列、不能改列类型、不能重命名、没有版本记录和回滚——这些限制现在就该写进文档，而不是留到真正需要那些操作时才被撞见。

### 决策：`AutoMigrate` 为准，删掉死的 SQL 文件

`deployments/migrations/001_create_todos.sql` 已经不一致（缺 `deleted_at`）且从未执行，继续留着只会让下一个读到它的人误以为它是权威定义。今天把它删掉，架构文档里写明 `model.Todo` 的 struct tag 是唯一的 schema 来源、`AutoMigrate` 是唯一的建表方式，并列出这条路线的已知限制（不处理删除/重命名/类型变更、无版本记录、无回滚）。

`deployments/` 目录只留 `docker-compose.yml`。

这条决策关闭 `docs/issues-backlog.md` C2。

---

## 第三件事：`User` 表设计

Day 23 要实现注册登录，今天先把表结构定下来，减少 Day 23 开工前的决策负担。

```go
// internal/model/user.go（Day 23 创建）
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"size:255;not null;uniqueIndex" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
```

几个决定点：

- **`Email` 加 `uniqueIndex`**：注册时的邮箱唯一性校验，用数据库约束兜底，不能只靠应用层查重（并发注册时应用层查重会有竞态窗口）。
- **`PasswordHash` 用 `json:"-"`**：和 `Todo.DeletedAt` 一样的模式——绝不能出现在任何 API 响应里，Day 21 学到的「面向客户端的信息和面向自己的信息是两回事」同样适用于字段级别，不止适用于错误信息。
- **`User` 不带软删除**：`Todo` 有 `gorm.DeletedAt` 是因为 TODO 项会被用户随手删除、可能需要恢复；账号删除是更重的操作（Day 23-25 都不涉及账号删除功能），现在不预留这个字段，需要时再加。
- **`Todo` 和 `User` 的关联**：架构文档「核心模型规划」里 TODO 最终形态已经写了 `UserID uint`，但**今天不在 `model.Todo` 里加这个字段**——外键关联应该和 JWT 认证一起落地（Day 25），因为在那之前没有「当前登录用户」的概念，加了 `UserID` 也没有值可写。今天只在文档里确认这个字段属于 Day 25 的范围，不是遗漏。

---

## 今日文档改动

`docs/architecture/todo-api.md` 需要改的地方：

1. 「核心模型规划」里 TODO 的 `Status string` 行改成注明「已决策保留 `Done bool`，见下方决策记录」，不要直接删掉这行——保留「曾经规划过什么、后来为什么变了」比看起来干净更有用。
2. 「已实现的模型」补一句：`deployments/migrations/001_create_todos.sql` 已删除，`AutoMigrate` + `model.Todo` struct tag 是唯一 schema 来源。
3. 新增一节「Day 22 设计决策」，把上面两条决策的现状、两条路线代价、结论都写进去（比这份教案更精简，面向未来读者而不是面向今天的学习过程）。
4. 「核心模型规划」的 `User` 小节从「尚未实现」的占位表格换成上面这份带决策理由的设计，并注明 `UserID` 外键留到 Day 25。
5. `docs/issues-backlog.md` 的状态表：D3 改成「已决策 (Day 22)」，C2 改成「已修 (Day 22)」，并在「已修」小节补一行。

---

## 今日验收标准

1. `docs/architecture/todo-api.md` 不再有「待决策」字样，`Status` vs `Done` 和 schema 权威来源都有明确结论和理由。
2. `User` 表设计写进文档，字段、约束、和 `Todo` 的关联时间点都说清楚。
3. `deployments/migrations/001_create_todos.sql` 删除（或者明确注明为历史留档，如果你更倾向保留但降级说明——两种做法都可以，选哪种今天的答案里说清楚就行）。
4. `docs/issues-backlog.md` D3、C2 状态更新。
5. 没有新代码，所以不需要 `make test`，但改完后跑一次 `make fmt vet` 确认没有破坏现有内容的引用。

## 可选挑战题

1. 给 `docs/architecture/todo-api.md` 补一张 ER 图（ASCII 即可），画出 `User` 1:N `Todo` 的关系，并标注这条外键还没有物理落地。
2. `AutoMigrate` 不处理删除列——如果今天决定保留 `Done bool` 但以后真的要迁到 `Status string`，写一段迁移步骤草稿（不用实现，只写步骤），验证一下「先加列、双写、切读、再删列」这种四步迁移在 GORM + `AutoMigrate` 组合下具体要怎么操作。

## 今天最容易踩的坑

1. **把「决策」写成「结论」，丢了理由。** 三周后回头看这份文档时，「为什么选 `Done bool`」比「选了 `Done bool`」更有价值——没有理由的决策记录等于没记录，下次有人想改就没有反驳的依据。
2. **删迁移脚本时不做备注，让人以为这是疏漏而不是决定。** 用 `git log` 能找到这个文件曾经存在，但 `docs/architecture/todo-api.md` 里如果不写一句「已删除，原因是 XXX」，读代码的人不会知道这是有意为之。
3. **把 `UserID` 现在就加到 `model.Todo` 里。** 没有认证系统之前这个字段永远是零值，加了它只是制造一个「看起来有关联但实际没用」的假象，不如明确写「Day 25 加」。
