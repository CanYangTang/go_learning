# Day 24 教案：TODO 完整 CRUD

## 学习目标

学完今天，需要能够做到：

1. 实现 `GET /api/v1/todos/:id`：按 ID 查询单个 TODO，不存在返回 `404`。
2. 实现 `PUT /api/v1/todos/:id`：全量更新一个 TODO，不存在返回 `404`。
3. 实现 `DELETE /api/v1/todos/:id`：删除一个 TODO，不存在返回 `404`。
4. 把路径参数 `:id` 的解析和校验放在正确的层（handler），非法 ID 返回 `400` 而不是 `500`。
5. 说清 GORM 的 `Save` 在「按 ID 更新一条不存在的记录」时会**变成插入**（upsert），以及为什么这对 PUT 语义是个坑。
6. 把 Day 20 定下的三层边界、Day 21 的错误处理规范、Day 23 的响应结构套到这三个新接口上——今天没有新架构决策。

今天**只做核心 CRUD**（GET by id / PUT / DELETE）。分页和状态筛选是 Issue #24 的可选挑战题，放在最后，不是必做。

---

## Day 24 的位置

Week 4 第三天。Day 23 加了 `user` 这条新链路，今天回到 `todo`，把它从「只能 Create / List」补成完整的 CRUD。

好消息是 `internal/repository/todo.go` 里 `FindByID`、`Update`、`Delete` 三个方法 Day 18 就写好了，一直没有调用方（`docs/issues-backlog.md` B3 记录了这一点）。今天正是给它们接上调用方的时候——但接的过程中会发现其中一个方法的实现对 PUT 语义有坑，需要重新审视，这是今天唯一需要动脑的地方。

---

## 第一件事：路径参数 `:id` 怎么解析、在哪一层校验

三个接口都长这样：`/api/v1/todos/:id`。`:id` 是 URL 里的一段字符串，Gin 用 `c.Param("id")` 取出来，拿到的是 `string`，要转成 `uint`。

实测 `strconv.ParseUint` 的边界行为（临时程序验证，跑完即删）：

```go
id, err := strconv.ParseUint("abc", 10, 64)
// id=0, err="strconv.ParseUint: parsing \"abc\": invalid syntax"

id, err := strconv.ParseUint("999999999999999999999", 10, 64)
// id=18446744073709551615, err="strconv.ParseUint: ... value out of range"
```

非数字和溢出都会返回 error。**这个解析和校验属于 handler 层**——它是「HTTP 请求怎么表达一个 ID」的细节，不是业务规则。service 层拿到的应该已经是一个干净的 `uint`，不需要知道 ID 是从 URL 路径里来的还是从别处来的。

handler 里的处理方式，和 Day 21 的绑定错误同一个模式：解析失败返回固定文案的 `400`，不要把 `strconv` 的原文透给客户端。

```go
func (h *TodoHandler) GetTodo(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		writeError(c, apperror.Validation("invalid id"))
		return
	}
	// ... 调 service
}

// parseID 是 handler 包内的小helper，三个接口共用。
func parseID(s string) (uint, error) {
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
```

---

## 第二件事：`GET /api/v1/todos/:id`

### repository

`FindByID` 已经写好了，Day 18 就把 `gorm.ErrRecordNotFound` 翻译成了 `(nil, nil)`：

```go
func (r *TodoRepository) FindByID(id uint) (*model.Todo, error) {
	var todo model.Todo
	err := r.db.First(&todo, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &todo, nil
}
```

（这里 `err == gorm.ErrRecordNotFound` 可以顺手改成 `errors.Is(err, gorm.ErrRecordNotFound)`，和 `user.go` 里的写法保持一致——直接 `==` 在 error 被 wrap 后会失效，`errors.Is` 更稳。今天改不改都行，改的话把 `todo.go` 和 `user.go` 统一了。）

### service

`service.TodoRepository` 接口现在只声明了 `Create`、`FindAll` 两个方法。今天要把 `FindByID`、`Update`、`Delete` 加进去——接口由使用方声明，service 用到几个就加几个。

```go
type TodoRepository interface {
	Create(todo *model.Todo) error
	FindAll() ([]model.Todo, error)
	FindByID(id uint) (*model.Todo, error)
	Update(todo *model.Todo) error
	Delete(id uint) error
}

func (s *TodoService) GetTodo(id uint) (*model.Todo, error) {
	todo, err := s.repo.FindByID(id)
	if err != nil {
		return nil, apperror.Internal("get todo failed")
	}
	if todo == nil {
		return nil, apperror.NotFound("todo not found")
	}
	return todo, nil
}
```

`repo.FindByID` 返回 `(nil, nil)` 表示「查询本身成功，但没这条记录」——service 把它翻译成 `apperror.NotFound`。这是 `apperror.NotFound` 第一次被业务代码用上（Day 21 加了这个构造函数，但当时只有 `NoRoute` 用）。

### handler

```go
func (h *TodoHandler) GetTodo(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		writeError(c, apperror.Validation("invalid id"))
		return
	}

	todo, err := h.service.GetTodo(id)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Body{Data: newTodoResponse(*todo), Message: "ok"})
}
```

`newTodoResponse` 已经有了（`ListTodos` 在用），直接复用。

---

## 第三件事：`PUT /api/v1/todos/:id`——今天唯一的坑

### GORM `Save` 的 upsert 行为（实测）

`internal/repository/todo.go` 里现成的 `Update` 是这样的：

```go
func (r *TodoRepository) Update(todo *model.Todo) error {
	return r.db.Save(todo).Error
}
```

看起来没问题，但 `Save` 有一个容易忽略的行为。实测（临时建表验证，跑完即删）：

```go
// 数据库里没有 id=42 这条记录
t := &TmpCrud{ID: 42, Title: "ghost"}
db.Save(t)
// err=nil, RowsAffected=1
// 再查：id=42 这条记录被创建出来了！
```

**`Save` 在主键对应的记录不存在时，会插入一条新记录**（带上你指定的那个 ID），而不是报错。对 PUT `/todos/42` 来说，如果 42 不存在，语义应该是 `404 Not Found`，绝不能是「悄悄创建一条 id=42 的记录」。

对比一下 `Delete` 和 `Updates` 的实测行为：

```go
db.Delete(&TmpCrud{}, 999999)
// err=nil, RowsAffected=0   ← 不报错，但影响 0 行

db.Model(&TmpCrud{}).Where("id = ?", 888888).Updates(map[string]any{"title": "x"})
// err=nil, RowsAffected=0   ← 同样，不存在就是影响 0 行
```

`Delete` 和 `Updates` 对不存在的记录是「影响 0 行、不报错」，可以靠 `RowsAffected == 0` 判断出「没这条记录」；`Save` 却会 upsert。

### 今天的处理方式：先查存在性，再更新

最直接、也最贴合现有代码的做法：service 层先 `FindByID` 确认记录存在（顺便这也是要返回给客户端的对象），存在才更新。

```go
func (s *TodoService) UpdateTodo(id uint, title string, done bool) (*model.Todo, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, apperror.Validation("title is required")
	}

	todo, err := s.repo.FindByID(id)
	if err != nil {
		return nil, apperror.Internal("update todo failed")
	}
	if todo == nil {
		return nil, apperror.NotFound("todo not found")
	}

	todo.Title = title
	todo.Done = done
	if err := s.repo.Update(todo); err != nil {
		return nil, apperror.Internal("update todo failed")
	}
	return todo, nil
}
```

因为 `FindByID` 已经确认这条记录存在，`repo.Update` 里的 `Save` 这时候走的是真正的 UPDATE 分支，upsert 行为不会被触发。

> 有一个理论上的 TOCTOU 窗口：`FindByID` 和 `Save` 之间，另一个请求可能把这条记录删了，于是 `Save` 又把它 upsert 回来。对一个单用户学习项目这个窗口可以接受，但值得知道它存在。更严格的写法是让 repository 的 `Update` 改用 `Model().Where("id = ?", id).Updates(...)` 并检查 `RowsAffected`，把「存在性判断」和「更新」合成一个原子操作——今天不强制改，但如果你想改，把理由写进注释。

### handler

```go
type UpdateTodoRequest struct {
	Title string `json:"title" binding:"required"`
	Done  bool   `json:"done"`
}

func (h *TodoHandler) UpdateTodo(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		writeError(c, apperror.Validation("invalid id"))
		return
	}

	var req UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("request_id=%s bind_error=%v", middleware.RequestIDFromContext(c), err)
		writeError(c, apperror.Validation("invalid request body"))
		return
	}

	todo, err := h.service.UpdateTodo(id, req.Title, req.Done)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Body{Data: newTodoResponse(*todo), Message: "ok"})
}
```

`Done` 字段没有 `binding:"required"`——`bool` 的 `required` 会把 `false` 当成「未提供」而拒掉，这是 Gin validator 的一个已知陷阱。`Done` 本来就允许是 `false`，所以不加 `required`，缺省时就是零值 `false`。

---

## 第四件事：`DELETE /api/v1/todos/:id`

### repository：用 `RowsAffected` 区分「删了」和「本来就没有」

现成的 `Delete` 不返回「删了几行」：

```go
func (r *TodoRepository) Delete(id uint) error {
	return r.db.Delete(&model.Todo{}, id).Error
}
```

如上一节实测，`Delete` 一个不存在的 ID 不会报错、`RowsAffected=0`。如果照这个实现，DELETE 一个不存在的 `/todos/999` 会返回 `200`——但语义上应该是 `404`。今天有两个选择：

**选择 A：service 层先 `FindByID` 判断存在性**，和 `UpdateTodo` 对称，不改 repository。
**选择 B：repository 的 `Delete` 返回 `RowsAffected`**，service 靠它判断。

```go
// 选择 B 的 repository 改法
func (r *TodoRepository) Delete(id uint) (int64, error) {
	result := r.db.Delete(&model.Todo{}, id)
	return result.RowsAffected, result.Error
}
```

两种都成立。选择 A 和 `UpdateTodo` 的写法一致、repository 签名不用改，代价是多一次查询；选择 B 少一次查询、更贴近 SQL 的真实语义，代价是改了 `Delete` 的签名（它现在没有调用方，改起来没成本）。**今天选哪个都行，选了就在代码注释里写一句为什么。** 下面按选择 A 写 service，因为它和 `UpdateTodo` 对称、这一天读起来一致：

```go
func (s *TodoService) DeleteTodo(id uint) error {
	todo, err := s.repo.FindByID(id)
	if err != nil {
		return apperror.Internal("delete todo failed")
	}
	if todo == nil {
		return apperror.NotFound("todo not found")
	}
	if err := s.repo.Delete(id); err != nil {
		return apperror.Internal("delete todo failed")
	}
	return nil
}
```

### handler：删除成功返回什么

```go
func (h *TodoHandler) DeleteTodo(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		writeError(c, apperror.Validation("invalid id"))
		return
	}

	if err := h.service.DeleteTodo(id); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Body{Data: nil, Message: "ok"})
}
```

删除成功用 `200` + 统一信封。**注意实际 body 是 `{"message":"ok"}`，不是 `{"data":null,"message":"ok"}`**——`response.Body.Data` 带 `json:"data,omitempty"`，`Data: nil` 会被整个省掉。（`204 No Content` 也是常见选择，但它要求响应体为空，和我们「所有成功响应都走统一信封」的约定冲突——为了一致性，这里用 `200` + 信封。这个取舍值得在 API 文档里写一句。）

> Day 24 实施时选了**选择 B**：`repo.Delete` 改成返回 `(int64, error)`，`DeleteTodo` 靠 `RowsAffected == 0` 判 `404`，不再多查一次 `FindByID`。端到端实测确认：删除一条已软删除的记录，`RowsAffected` 也是 0（软删除行被 GORM 过滤掉了），所以重复 DELETE 同一 id 正确返回 `404`。

---

## 第五件事：软删除的既有行为

`model.Todo` 有 `DeletedAt gorm.DeletedAt`（Day 18 加的软删除）。这意味着 `r.db.Delete(...)` 执行的不是物理 `DELETE`，而是 `UPDATE todos SET deleted_at = NOW()`。后续所有 GORM 查询（`First`、`Find`）默认自动带上 `WHERE deleted_at IS NULL`，所以：

- 删掉一条 TODO 后，`GET /todos/:id` 会返回 `404`（查不到）。
- 删掉一条 TODO 后，`GET /todos` 列表里不再出现它。

这些是 GORM 软删除的既有行为，今天不需要额外写代码，但**验收时要实测确认**——删除后再查确实是 404、列表确实少了一条。这也是今天 `RowsAffected`/`FindByID` 判断能正常工作的前提：软删除后 `FindByID` 走的还是「查不到 → `(nil, nil)`」这条路。

---

## 第六件事：路由

```go
v1 := router.Group("/api/v1")
{
	v1.GET("/health", handler.HealthHandler)
	v1.POST("/todos", todoHandler.CreateTodo)
	v1.GET("/todos", todoHandler.ListTodos)
	v1.GET("/todos/:id", todoHandler.GetTodo)
	v1.PUT("/todos/:id", todoHandler.UpdateTodo)
	v1.DELETE("/todos/:id", todoHandler.DeleteTodo)
	v1.POST("/users/register", userHandler.Register)
	v1.POST("/users/login", userHandler.Login)
}
```

不用改中间件、不用改装配链（`TodoHandler` 已经装配好了，只是多挂几个方法）。

---

## 今日代码结构

改动（今天没有新文件）：

```
internal/repository/todo.go   # 可选：FindByID 的 == 改 errors.Is；Delete 若选方案 B 改签名
internal/service/todo.go      # TodoRepository 接口加 3 个方法；新增 GetTodo/UpdateTodo/DeleteTodo
internal/handler/todo.go      # TodoService 接口加 3 个方法；新增 parseID/GetTodo/UpdateTodo/DeleteTodo；UpdateTodoRequest
cmd/server/main.go            # 注册 3 个新路由
```

加测试（沿用现有模式）：

```
internal/service/todo_test.go        # 给 fakeTodoRepo 补 FindByID/Update/Delete，加 GetTodo/UpdateTodo/DeleteTodo 用例（含 404 分支）
internal/handler/todo_test.go        # 给 fakeTodoService 补 3 个方法，加 3 个接口的用例（含非法 id 的 400、404）
internal/repository/todo_test.go     # 已有 FindByID/Update/Delete 的集成用例，确认软删除后 FindByID 返回 nil
```

---

## 今日验收标准

1. `GET /api/v1/todos/:id`：存在返回 `200` + 单个 todo；不存在返回 `404`；非数字 id 返回 `400`。
2. `PUT /api/v1/todos/:id`：存在则更新并返回 `200` + 更新后的 todo；不存在返回 `404`（**实测确认不会 upsert 出一条新记录**）；`title` 为空返回 `400`。
3. `DELETE /api/v1/todos/:id`：存在则删除返回 `200`；不存在返回 `404`；删除后再 `GET` 同一 id 返回 `404`，且列表里不再出现。
4. 三层边界不破：service/handler 不 import `gorm`；`:id` 解析在 handler，业务判断在 service。
5. `make fmt test vet` 全部通过。
6. `docs/api/todo-api.md` 把这三个接口从「计划中」挪到「已实现」，示例从运行中的服务上抓。

## 可选挑战题（Issue #24 原文）

1. **分页**：`GET /api/v1/todos` 加 `page`/`page_size` 查询参数，响应形状改成 `{ "data": { "items": [], "page", "page_size", "total" }, "message": "ok" }`。注意这是对现有 `GET /todos` 的**破坏性改动**（现在直接返回数组），`docs/api/todo-api.md`「计划中」小节已经写了目标形状，实现时要同步文档，并想清楚 `total` 怎么在软删除存在时算对。
2. **状态筛选**：加 `?done=true`/`?done=false` 筛选。想清楚「不传 `done`」和「`done=false`」是两回事——前者返回全部，后者只返回未完成的。

## 今天最容易踩的坑

1. **PUT 一个不存在的 id 时 `Save` 悄悄 upsert 出新记录。** 今天唯一的实测坑，验收标准第 2 条专门测。必须先 `FindByID` 判存在性（或改用 `Updates` + `RowsAffected`）。
2. **`:id` 解析失败返回 500 而不是 400。** `strconv.ParseUint` 的 error 是客户端输入问题，走 `apperror.Validation`，和 Day 23 的 bcrypt 72 字节坑同一类。
3. **`UpdateTodoRequest.Done` 加了 `binding:"required"`。** `bool` 字段加 `required`，Gin 会把合法的 `false` 当成「缺失」拒掉。`Done` 不加 `required`。
4. **DELETE 不存在的 id 返回 200。** `db.Delete` 对不存在的 id 不报错，不判断就会返回成功。要么 service 先查，要么 repository 返回 `RowsAffected`。
5. **忘了软删除的既有行为，以为删除后记录还能查到。** 删除是 `deleted_at` 打标记，GORM 查询自动过滤，删除后 `FindByID` 走的是 `(nil, nil)` 分支——验收时实测确认这条链路。
