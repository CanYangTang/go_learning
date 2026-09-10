# TODO API 架构说明

> 「规划」小节写的是最终形态，「当前状态」小节写的是 `main` 分支上真实存在的东西。
> 最后核对：2026-09-08（Day 21）

## 项目目标

TODO API 是本学习项目的最终产出，用于串联 Go 基础、并发、HTTP、Gin、GORM、MySQL、JWT、测试和部署准备。

## 技术栈路线

| 阶段 | 技术重点 |
|------|----------|
| Week 1 | Go 基础、函数、错误处理、包管理、测试 |
| Week 2 | `net/http`、JSON、goroutine、channel、文件操作 |
| Week 3 | Gin、GORM、MySQL、中间件、分层架构 |
| Week 4 | JWT、完整 CRUD、测试、文档、部署准备 |

## 目标分层

```text
request
  ↓
middleware
  ↓
handler
  ↓
service
  ↓
repository
  ↓
database
```

### handler

- 解析 HTTP 请求。
- 绑定和校验请求参数。
- 调用 service。
- 返回统一 JSON 响应。

### service

- 编排业务逻辑。
- 处理业务规则。
- 不直接依赖 HTTP 细节。

### repository

- 封装数据访问。
- 已切换到 GORM + MySQL，不再保留内存实现。

### model

- 定义 User、TODO 等核心数据模型。
- 带 GORM tag 和 JSON tag。

### middleware

- 日志。
- CORS。
- JWT 认证。
- 请求 ID。
- panic 恢复。

## 当前状态（Day 21）

### 已落地的分层

```text
request
  ↓ RequestID → Logging → Recovery → CORS → AuthPlaceholder
handler   internal/handler   TodoHandler / HealthHandler / NotFoundHandler
  ↓
service   internal/service   TodoService
  ↓
repository internal/repository TodoRepository（GORM）
  ↓
MySQL
```

中间件挂载顺序是有讲究的，不能随手调：`Recovery` 必须在 `Logging` **之内**，否则 panic 被兜住时控制流不会回到 `Logging` 的 `c.Next()` 之后，那次请求就没有访问日志；`RequestID` 在最外层，`Recovery` 才能在日志里带上 request ID；`CORS` 在 `Recovery` 之内，500 响应才保留跨域头。

### 已实现的模型

只有 TODO，`User` 还没写（Day 25）。实际字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | uint | 主键 |
| Title | string | 标题，`size:255;not null` |
| Done | bool | 是否完成，`default:false` |
| CreatedAt | time.Time | 创建时间 |
| UpdatedAt | time.Time | 更新时间 |
| DeletedAt | gorm.DeletedAt | 软删除标记，带索引，JSON 中隐藏 |

表结构由 `db.AutoMigrate(&model.Todo{})` 在启动时创建，`deployments/migrations/001_create_todos.sql` 目前是另一份**不一致且从未执行**的定义（`docs/issues-backlog.md` C2，Day 22 处理）。

### 三条设计决策（Day 20 确立）

**1. 接口由使用方声明，不由实现方声明。**

`service.TodoRepository` 只列出 service 真正用到的 2 个方法，而不是 repository 的全部 6 个；`handler.TodoService` 同理。Go 的接口是隐式满足的，所以不需要实现方 import 使用方。好处是每层的依赖面就是它自己声明的那几个方法，测试时的假实现也只需要实现那几个。

**2. 依赖注入只发生在 `cmd/server/main.go`。**

`NewTodoRepository(db)` → `NewTodoService(repo)` → `NewTodoHandler(svc)` 这条装配链只在 main 里出现一次。任何层都不自己 `new` 下游依赖，也不读全局变量。

**3. 错误只有一个出口。**

业务层返回 `apperror.Error`（值类型，`Code` / `Message` / `StatusCode`），handler 里统一由 `writeError` 用 `errors.As` 还原成 HTTP 状态码和错误信封。handler 里不出现第二处 `c.JSON(http.StatusBadRequest, ...)`。`writeError` 是 unexported 的，所以 `NotFoundHandler` 放在 `internal/handler` 包内才能复用它；`internal/middleware/recovery.go` 拿不到它，只能用 `pkg/response` 重建同样形状的信封。

### 边界约束（可用 import 检查）

- `internal/service` 既不 import `gin` 也不 import `gorm`。
- `internal/handler` 不 import `gorm`。
- 依赖方向单向向下，下层不 import 上层。

## 核心模型规划

以下是最终形态，**尚未实现**。

### User（Day 25）

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | uint | 主键 |
| Email | string | 邮箱，唯一 |
| PasswordHash | string | 哈希后的密码 |
| CreatedAt | time.Time | 创建时间 |
| UpdatedAt | time.Time | 更新时间 |

### TODO（最终形态）

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | uint | 主键 |
| UserID | uint | 所属用户 |
| Title | string | 标题 |
| Description | string | 描述 |
| Status | string | 状态：pending/done |
| DueDate | *time.Time | 截止日期 |
| CreatedAt | time.Time | 创建时间 |
| UpdatedAt | time.Time | 更新时间 |

这里的 `Status string` 与当前实现的 `Done bool` 冲突，二选一的决策留到 Day 22（`docs/issues-backlog.md` D3）。

## 统一响应规划

成功响应：

```json
{
  "data": {},
  "message": "ok"
}
```

错误响应：

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "invalid request"
  }
}
```

## 完成标准

| 标准 | 状态 |
|------|------|
| 服务可以本地启动 | 已完成（需先起 MySQL） |
| 数据持久化到 MySQL | 已完成 |
| `go test ./...` 通过 | 已完成 |
| TODO 支持创建、查询 | 已完成 |
| TODO 支持更新、删除 | 待完成 |
| 列表支持分页和状态筛选 | 待完成 |
| 注册登录可用 | 待完成 |
| JWT 能保护 TODO 路由 | 待完成 |
| README 和 API 文档能指导别人运行项目 | 已完成（Day 21 重写） |
