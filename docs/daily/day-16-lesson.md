# Day 16 教案：TODO 请求绑定和 JSON 响应

## 学习目标

学完今天，需要能够做到：

1. 理解 Gin 的请求绑定（ShouldBindJSON、BindJSON）。
2. 实现 TODO 的创建和列表接口。
3. 使用 Gin 返回 JSON 响应。
4. 理解请求结构体和响应结构体的设计。
5. 编写 Gin handler 的测试。
6. 使用 `pkg/response` 统一响应格式。

---

## Day 16 的位置

Day 15 学了 Gin 框架和路由组。

Day 16 会在 `/api/v1` 路由组下实现 TODO 相关接口：

- `POST /api/v1/todos` — 创建 TODO
- `GET /api/v1/todos` — 获取 TODO 列表

---

## TODO 数据结构

定义 TODO 模型：

```go
type Todo struct {
    ID     uint   `json:"id"`
    Title  string `json:"title"`
    Done   bool   `json:"done"`
}
```

---

## 请求绑定

Gin 提供了多种请求绑定方法：

### ShouldBindJSON

```go
var req CreateTodoRequest
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
}
```

- 从请求体解析 JSON 到结构体。
- 解析失败返回错误，不会自动中止请求。
- 需要手动处理错误。

### BindJSON

```go
var req CreateTodoRequest
if err := c.BindJSON(&req); err != nil {
    // 自动返回 400 错误响应
    return
}
```

- 解析失败自动返回 400 错误。
- 更简洁，但错误格式不可自定义。

今天建议使用 `ShouldBindJSON`，便于自定义错误响应。

---

## 请求结构体

创建请求结构体：

```go
type CreateTodoRequest struct {
    Title string `json:"title" binding:"required"`
}
```

- `binding:"required"` 表示字段必填，Gin 会自动校验。

列表请求可以简单：

```go
type ListTodoRequest struct {
    // 暂时无参数
}
```

---

## 响应结构体

使用 `pkg/response` 的统一格式：

```go
// 成功响应
response.JSON(c, http.StatusOK, todo)

// 错误响应
response.Error(c, http.StatusBadRequest, "INVALID_TITLE", "title is required")
```

---

## 创建 TODO 接口

### 请求

```text
POST /api/v1/todos
Content-Type: application/json

{
    "title": "Learn Go"
}
```

### 响应

```json
{
    "message": "ok",
    "data": {
        "id": 1,
        "title": "Learn Go",
        "done": false
    }
}
```

---

## 列表 TODO 接口

### 请求

```text
GET /api/v1/todos
```

### 响应

```json
{
    "message": "ok",
    "data": [
        {
            "id": 1,
            "title": "Learn Go",
            "done": false
        },
        {
            "id": 2,
            "title": "Write tests",
            "done": true
        }
    ]
}
```

---

## Handler 实现

### CreateTodo

```go
func (h *TodoHandler) CreateTodo(c *gin.Context) {
    var req CreateTodoRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
        return
    }

    todo := Todo{
        Title: req.Title,
        Done:  false,
    }

    // TODO: 保存到数据库（Day 17 实现）
    // 这里暂时用内存存储

    response.JSON(c, http.StatusCreated, todo)
}
```

### ListTodos

```go
func (h *TodoHandler) ListTodos(c *gin.Context) {
    // TODO: 从数据库查询（Day 17 实现）
    // 这里暂时返回空列表

    todos := []Todo{}

    response.JSON(c, http.StatusOK, todos)
}
```

---

## 今日代码结构

```
internal/handler/todo.go       # TODO handler
internal/handler/todo_test.go  # TODO handler 测试
```

---

## Handler 结构体

使用结构体方法，便于后续注入依赖：

```go
type TodoHandler struct {
    // TODO: 后续注入 service 或 repository
}

func NewTodoHandler() *TodoHandler {
    return &TodoHandler{}
}

func (h *TodoHandler) CreateTodo(c *gin.Context) {
    // ...
}

func (h *TodoHandler) ListTodos(c *gin.Context) {
    // ...
}
```

---

## 路由注册

在 `cmd/server/main.go` 中注册路由：

```go
todoHandler := handler.NewTodoHandler()

v1 := router.Group("/api/v1")
{
    v1.GET("/health", handler.HealthHandler)
    v1.POST("/todos", todoHandler.CreateTodo)
    v1.GET("/todos", todoHandler.ListTodos)
}
```

---

## 测试设计

至少写 3 组测试：

1. `TestCreateTodoSuccess` — 创建成功。
2. `TestCreateTodoInvalidJSON` — 无效 JSON 返回 400。
3. `TestListTodosEmpty` — 空列表返回空数组。

示例：

```go
func TestCreateTodoSuccess(t *testing.T) {
    gin.SetMode(gin.TestMode)

    h := NewTodoHandler()
    router := gin.New()
    router.POST("/api/v1/todos", h.CreateTodo)

    body := `{"title":"Learn Go"}`
    req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    recorder := httptest.NewRecorder()

    router.ServeHTTP(recorder, req)

    if recorder.Code != http.StatusCreated {
        t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
    }
}
```

---

## 今日验收标准

完成后应该满足：

1. `internal/handler/todo.go` 实现 `TodoHandler`、`CreateTodo`、`ListTodos`。
2. `POST /api/v1/todos` 能创建 TODO 并返回 201。
3. `GET /api/v1/todos` 能返回 TODO 列表。
4. 无效 JSON 请求返回 400 和错误信息。
5. 至少写 3 组测试。
6. `go test ./internal/handler` 通过。
7. `make test` 通过。
8. 能解释 `ShouldBindJSON`、请求结构体、响应结构体。

---

## 可选挑战题：请求参数验证

Issue 的可选挑战是增加请求参数验证。

Gin 内置了 validator，可以在结构体 tag 中添加验证规则：

```go
type CreateTodoRequest struct {
    Title string `json:"title" binding:"required,min=1,max=100"`
}
```

- `required`：必填。
- `min=1`：最小长度 1。
- `max=100`：最大长度 100。

---

## 今天最容易踩的坑

### 坑 1：忘记 `Content-Type`

测试时要设置：

```go
req.Header.Set("Content-Type", "application/json")
```

否则 `ShouldBindJSON` 无法解析。

---

### 坑 2：响应状态码错误

创建成功应该返回 `http.StatusCreated`（201），不是 `http.StatusOK`（200）。

---

### 坑 3：忘记 `binding` tag

请求结构体的字段要加 `json` 和 `binding` tag：

```go
Title string `json:"title" binding:"required"`
```

否则 `ShouldBindJSON` 不会校验必填。

---

### 坑 4：空列表返回 `null` 而不是 `[]`

空 slice 返回 JSON 时：

```go
todos := []Todo{}  // 空切片，JSON 会是 []
todos := nil       // nil，JSON 可能是 null
```

建议用 `[]Todo{}` 确保返回空数组。

---

### 坑 5：Handler 方法签名错误

Handler 方法必须是 `func(c *gin.Context)`，接收者是 `(h *TodoHandler)`：

```go
func (h *TodoHandler) CreateTodo(c *gin.Context)  // 正确
func CreateTodo(c *gin.Context)                    // 普通函数，也可以
```

两种都可以，但使用结构体方法便于后续注入依赖。