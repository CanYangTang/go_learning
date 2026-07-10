# Day 16 学习记录

## 日期

2026-07-09

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/16

## 今日教案

- 教案文档：`docs/daily/day-16-lesson.md`

## 核心任务

- 学习 Gin 的请求绑定（ShouldBindJSON）。
- 实现 TODO 的创建和列表接口。
- 使用 `pkg/response` 统一响应格式。
- 理解请求结构体和响应结构体的设计。

## 验收标准

- `internal/handler/todo.go` 实现 `TodoHandler`、`CreateTodo`、`ListTodos`。
- `POST /api/v1/todos` 能创建 TODO 并返回 201。
- `GET /api/v1/todos` 能返回 TODO 列表。
- 无效 JSON 请求返回 400 和错误信息。
- 至少写 3 组测试。
- `go test ./internal/handler` 通过。
- `make test` 通过。
- 能解释 `ShouldBindJSON`、请求结构体、响应结构体。

## 可选挑战题

- 增加请求参数验证（`binding:"required,min=1,max=100"`）。

## 答疑记录

- 待记录。

## 今日产出

- 实现 `internal/handler/todo.go`：`TodoHandler`、`CreateTodo`、`ListTodos`。
- 使用 `c.ShouldBindJSON` 绑定请求体，配合 `binding:"required"` 校验必填。
- 创建成功返回 201，使用 `response.Body` 统一响应格式。
- 列表接口返回空切片 `[]Todo{}`，确保 JSON 序列化为空数组。
- 编写测试覆盖成功创建、无效 JSON、缺少必填字段、空列表。

## 运行过的命令

```bash
go test ./internal/handler
make fmt
make test
make vet
```

## 代码 Review 结论

- `TodoHandler` 使用结构体方法，便于后续注入 service 或 repository。
- `CreateTodoRequest` 的 `Title` 字段使用 `binding:"required"` 标签，Gin 自动校验必填。
- `ShouldBindJSON` 解析失败时返回错误，需要手动处理并返回 400 响应。
- 使用 `response.Body` 和 `response.ErrorBody` 保证响应格式统一。
- 创建成功返回 `http.StatusCreated`（201），符合 RESTful 规范。
- 空列表用 `[]Todo{}` 而不是 `nil`，确保 JSON 输出 `"data":[]` 而不是 `"data":null`。
- 测试使用 `strings.Contains` 验证响应内容，足够简单有效。
- `go test ./internal/handler`、`make fmt`、`make test`、`make vet` 均已通过。

## 今日小测试

1. `ShouldBindJSON` 和 `BindJSON` 有什么区别？为什么推荐用 `ShouldBindJSON`？
   - 回答：`BindJSON` 失败时会偷偷调用 `c.AbortWithStatus(400)`，如果再手动写响应，就会写两次，导致警告或行为异常。`ShouldBindJSON` 只做绑定，不碰响应，更可控。
   - 结果：通过。
   - 标准答案：`BindJSON` 解析失败时自动返回 400 并终止请求，无法自定义错误格式；`ShouldBindJSON` 只返回错误，由调用者决定如何响应，灵活性更高。
2. `binding:"required"` 标签有什么作用？如果去掉会怎样？
   - 回答：`ShouldBindJSON` 解析时会验证字段：有 `required` 时字段缺失或为空字符串返回 `error`；没有 `required` 时字段缺失不报错，用零值填充。
   - 结果：通过。
   - 标准答案：`binding:"required"` 表示字段必填，Gin 的 validator 会校验，缺失时 `ShouldBindJSON` 返回错误。去掉后字段变为可选，缺失时使用零值。
3. 为什么要用 `[]Todo{}` 而不是 `nil` 来表示空列表？
   - 回答：核心原因是 JSON 序列化结果不同。返回给前端 `null` 会导致错误，返回 `[]` 前端可以正常遍历，不会报错。
   - 结果：通过。
   - 标准答案：`nil` slice 序列化为 `null`，`[]Todo{}` 序列化为 `[]`。前端遍历 `null` 会报错，遍历空数组则正常。RESTful API 应返回空数组而非 `null`。
4. `response.Body` 结构体的作用是什么？为什么要统一响应格式？
   - 回答：统一响应格式。前端可以写统一的响应拦截器，不用每个接口单独处理。错误和成功格式一致，便于调试。规范团队协作，所有接口风格统一。
   - 结果：通过。
   - 标准答案：统一响应格式让前端可以写通用的响应处理逻辑，成功和错误结构一致，降低集成成本，便于调试和文档化。
5. 创建接口为什么返回 201 而不是 200？
   - 回答：201 明确告诉客户端不只是"请求成功"，而是"新资源已经被创建了"。实际影响：REST API 规范要求；部分客户端/框架会根据 201 做特殊处理（比如自动跳转到新资源）；语义清晰，便于接口文档和调试。
   - 结果：通过。
   - 标准答案：201 Created 表示资源创建成功，语义更精确；200 OK 只表示请求成功。RESTful 规范建议 POST 创建资源返回 201。

## 测试结果

- 5 题全部通过。

## 遇到的问题

- `pkg/response` 原本为 `net/http` 设计，Gin 中需要直接用 `c.JSON()` 配合 `response.Body` 结构体。

## 关键收获

1. `ShouldBindJSON` 解析失败返回错误，可以自定义错误响应格式；`BindJSON` 会自动返回 400。
2. `binding:"required"` 让 Gin 自动校验必填字段，校验失败时 `ShouldBindJSON` 返回错误。
3. 空切片 `[]Todo{}` 序列化为 `[]`，而 `nil` 可能序列化为 `null`。
4. 统一响应格式让 API 消费方更容易处理响应。
5. POST 创建资源返回 201 Created，GET 列表返回 200 OK，符合 RESTful 规范。

## 明日计划

- 进入 Day 17：MySQL 连接和基础 CRUD。
