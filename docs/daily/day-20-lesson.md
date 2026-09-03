# Day 20 教案：分层架构

## 学习目标

学完今天，需要能够做到：

1. 说清 handler / service / repository 三层各自的职责边界。
2. 理解依赖方向，以及 Go 里「接口定义在调用方」的惯例。
3. 把现在假数据的 TODO 接口改成真正走 repository 的链路。
4. 用 `apperror` 让业务错误能被 handler 映射成 HTTP 状态码。
5. 在 `cmd/server/main.go` 里完成依赖组装（wiring）。
6. 为 service 写不依赖数据库的单元测试。

---

## Day 20 的位置

前面几天各层都零散地写过了：

- Day 15、16：`internal/handler/` 有了 Gin 路由和请求绑定。
- Day 17、18：`internal/config/`、`internal/model/`、`internal/repository/` 有了数据库和 GORM。
- Day 19：`internal/middleware/` 补齐了请求链路的横切能力。

但它们目前**没有连起来**。今天的任务不是加功能，而是把已有的代码接成一条完整链路。

---

## 现状的问题

看一眼 `internal/handler/todo.go`：

```go
type TodoHandler struct {
    // TODO: inject service or repository later
}

func (h *TodoHandler) ListTodos(c *gin.Context) {
    todos := []Todo{}
    c.JSON(http.StatusOK, response.Body{Data: todos, Message: "ok"})
}
```

三个问题：

1. `ListTodos` 永远返回空数组，`CreateTodo` 创建完就丢掉了，数据没有落库。
2. `internal/repository/todo.go` 已经能用，但没有任何地方调用它。
3. 业务规则（比如 title 该怎么校验、什么算合法）无处安放，只能塞在 handler 里。

第 3 点是引入 service 层的真正原因，不是为了「看起来更规范」。

---

## 三层职责

| 层 | 负责 | 不负责 |
|----|------|--------|
| handler | 解析 HTTP 请求、调用 service、把结果和错误写成 JSON | 业务规则、SQL |
| service | 业务规则、校验、编排多个 repository 调用、事务边界 | HTTP 细节、SQL 细节 |
| repository | 数据库读写 | 业务规则、HTTP |

判断代码该放哪一层，有两个简单的检查方法：

- service 里出现 `gin` 或 `http`，说明 HTTP 细节漏进来了。
- service 里出现 `gorm`，说明数据库细节漏进来了。

反过来，handler 里出现 `gorm`，说明中间那层被跳过了。

---

## 依赖方向

依赖只能单向往下：

```text
handler  ->  service  ->  repository  ->  model
```

`model` 是所有层都能引用的数据结构，所以它不依赖任何人。

注意一个反例：**repository 不能反过来引用 service**。一旦出现双向引用，Go 会直接报 import cycle。

---

## 接口该定义在哪一层

这是今天最重要的一个概念。

Java 那套习惯是「provider 定义接口」，即 repository 包里定义 `TodoRepository` 接口再自己实现。

Go 的惯例相反：**接口定义在调用方（consumer）**。

也就是说，`TodoRepository` 接口写在 `internal/service/` 里：

```go
// internal/service/todo.go
type TodoRepository interface {
    Create(todo *model.Todo) error
    FindAll() ([]model.Todo, error)
}
```

`repository.TodoRepository` 这个 struct 不需要写 `implements`，只要方法签名对得上，就自动满足这个接口。

这样做有三个好处：

1. 调用方声明「我需要什么」，而不是被动接受「你提供什么」。
2. 接口可以只声明**用得到的方法**。repository 有 6 个方法，service 现在只用 2 个，接口就只写 2 个。
3. 测试时随手写个假实现就能替换，不需要数据库。

---

## service 层怎么写

```go
package service

import (
    "strings"

    "github.com/CanYangTang/go_learning/internal/model"
    "github.com/CanYangTang/go_learning/pkg/apperror"
)

// TodoRepository is the persistence contract the service depends on.
type TodoRepository interface {
    Create(todo *model.Todo) error
    FindAll() ([]model.Todo, error)
}

// TodoService holds the business rules for todos.
type TodoService struct {
    repo TodoRepository
}

func NewTodoService(repo TodoRepository) *TodoService {
    return &TodoService{repo: repo}
}

func (s *TodoService) CreateTodo(title string) (*model.Todo, error) {
    title = strings.TrimSpace(title)
    if title == "" {
        return nil, apperror.Validation("title is required")
    }

    todo := &model.Todo{Title: title}
    if err := s.repo.Create(todo); err != nil {
        return nil, apperror.Internal("create todo failed")
    }
    return todo, nil
}

func (s *TodoService) ListTodos() ([]model.Todo, error) {
    todos, err := s.repo.FindAll()
    if err != nil {
        return nil, apperror.Internal("list todos failed")
    }
    if todos == nil {
        todos = []model.Todo{}
    }
    return todos, nil
}
```

三个要点：

- `strings.TrimSpace` 这类规则属于业务，放 service，不放 handler，也不放 repository。
- 底层错误要包成 `apperror`，不要把 `gorm` 的错误原样透出去，否则 handler 只能猜状态码。
- `FindAll` 可能返回 nil slice，归一化成空切片，JSON 才是 `[]` 而不是 `null`。

---

## handler 层怎么改

handler 也按「调用方定义接口」的规则，声明自己需要的 service：

```go
// TodoService is the business contract the handler depends on.
type TodoService interface {
    CreateTodo(title string) (*model.Todo, error)
    ListTodos() ([]model.Todo, error)
}

type TodoHandler struct {
    service TodoService
}

func NewTodoHandler(service TodoService) *TodoHandler {
    return &TodoHandler{service: service}
}
```

handler 里保留一个响应结构体，负责把 `model.Todo` 转成对外的 JSON 形状：

```go
func newTodoResponse(todo model.Todo) Todo {
    return Todo{ID: todo.ID, Title: todo.Title, Done: todo.Done}
}
```

这一步叫边界转换。好处是数据库字段变化不会直接改变 API 契约。

错误处理统一收口到一个小函数：

```go
func writeError(c *gin.Context, err error) {
    var appErr apperror.Error
    if errors.As(err, &appErr) {
        c.JSON(appErr.StatusCode, response.ErrorBody{
            Error: response.ErrorPayload{Code: appErr.Code, Message: appErr.Message},
        })
        return
    }

    c.JSON(http.StatusInternalServerError, response.ErrorBody{
        Error: response.ErrorPayload{Code: "INTERNAL_ERROR", Message: "internal error"},
    })
}
```

`apperror.Error` 自带 `StatusCode`，所以 handler 不需要 if-else 判断业务错误类型。

顺带修掉一个不一致：现在绑定失败返回的 code 是 `INVALID_REQUEST`，但 `docs/api/todo-api.md` 里定的是 `VALIDATION_ERROR`。今天统一走 `apperror.Validation`。

---

## 依赖组装

三层都不自己 new 依赖，全部在 `cmd/server/main.go` 里组装一次：

```go
db, err := config.ConnectGorm(dsn)
if err != nil {
    log.Fatal(err)
}

todoRepo := repository.NewTodoRepository(db)
todoService := service.NewTodoService(todoRepo)
todoHandler := handler.NewTodoHandler(todoService)

v1 := router.Group("/api/v1")
{
    v1.GET("/health", handler.HealthHandler)
    v1.POST("/todos", todoHandler.CreateTodo)
    v1.GET("/todos", todoHandler.ListTodos)
}
```

这种「只在入口处组装」的写法叫依赖注入。除了 main，其他地方都不知道数据库是怎么连上的。

`internal/config/database.go` 现在只提供 `*sql.DB`，repository 要的是 `*gorm.DB`，所以今天要补一个 GORM 版本的连接函数：

```go
// internal/config/gorm.go
func ConnectGorm(dsn string) (*gorm.DB, error) {
    return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}
```

一个副作用要提前知道：**今天之后 `make run` 需要先启动 MySQL**。

```bash
docker compose -f deployments/docker-compose.yml up -d
```

连不上就直接 `log.Fatal` 退出，而不是带着半残状态启动，这是服务端的常规做法。

---

## 测试策略

分层最直接的回报就是测试变简单了。

| 层 | 替身 | 需要数据库 |
|----|------|-----------|
| service | 假 repository | 不需要 |
| handler | 假 service | 不需要 |
| repository | 真 MySQL | 需要（Day 18 的集成测试） |

service 的假 repository 大概长这样：

```go
type fakeTodoRepo struct {
    todos     []model.Todo
    nextID    uint
    createErr error
}

func (f *fakeTodoRepo) Create(todo *model.Todo) error {
    if f.createErr != nil {
        return f.createErr
    }
    f.nextID++
    todo.ID = f.nextID
    f.todos = append(f.todos, *todo)
    return nil
}
```

因为 `TodoRepository` 接口只声明了 2 个方法，这个假实现也只需要写 2 个方法。这就是「接口只声明用得到的方法」的收益。

---

## 今日代码结构

```text
internal/service/
  todo.go
  todo_test.go
internal/handler/
  todo.go          # 改：依赖 service
  todo_test.go     # 改：用假 service
internal/config/
  gorm.go          # 新增
cmd/server/main.go # 改：组装依赖 + 注册 todos 路由
```

---

## 今日验收标准

1. `internal/service/` 里有 `TodoService`，包含业务校验和错误包装。
2. `TodoRepository` 接口定义在 service 包，不在 repository 包。
3. handler 通过接口依赖 service，不再自己造数据。
4. `cmd/server/main.go` 完成 repo → service → handler 的组装。
5. service 层单元测试不依赖数据库。
6. `make fmt`、`make test`、`make vet` 全部通过。
7. 能说清三层职责边界和依赖方向。

---

## 可选挑战题：service 单元测试

用假 repository 覆盖这几种情况：

- title 正常 → 创建成功，ID 被回填。
- title 为空或全是空格 → 返回 `VALIDATION_ERROR`，且 repository 一次都没被调用。
- repository 报错 → 被包成 `INTERNAL_ERROR`。
- 列表为空 → 返回空切片而不是 nil。

第二条尤其值得测：校验失败时不该白跑一次数据库。

---

## 今天最容易踩的坑

### 坑 1：接口定义在 repository 包

会导致 service 被迫依赖 repository 包，接口的解耦作用就没了，测试也换不掉实现。接口应该定义在调用方。

### 坑 2：service 里 import gin

一旦 service 收 `*gin.Context`，它就绑死在 HTTP 上，没法被定时任务或 CLI 复用，测试也得造 Gin 上下文。service 只接收和返回普通 Go 值。

### 坑 3：service 直接返回 gorm 错误

handler 拿到一个裸 error，只能统一返回 500，分不清是参数问题还是数据库问题。业务错误必须包成 `apperror`。

### 坑 4：handler 直接把 `model.Todo` 当响应体

数据库字段一改，API 契约就跟着变。中间加一层响应结构体转换。

### 坑 5：repository 返回 nil slice 直接透出

JSON 会变成 `"data":null`，前端要额外判空。service 里归一化成 `[]model.Todo{}`。

### 坑 6：在各层内部自己 new 依赖

比如 handler 里直接 `service.NewTodoService(...)`。这样就没法替换实现，测试也注入不进去。依赖统一在 main 组装。
