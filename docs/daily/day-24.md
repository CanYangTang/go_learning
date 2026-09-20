# Day 24 学习记录

## 日期

2026-09-16

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/24

## 今日教案

- 教案文档：`docs/daily/day-24-lesson.md`

## 核心任务

- 实现 `GET /api/v1/todos/:id`：按 ID 查询单个 TODO，不存在返回 `404`。
- 实现 `PUT /api/v1/todos/:id`：全量更新，不存在返回 `404`（**不能被 `Save` upsert 出新记录**）。
- 实现 `DELETE /api/v1/todos/:id`：删除，不存在返回 `404`；删除后再查确实是 404、列表里不再出现（软删除既有行为）。
- `:id` 解析放在 handler 层，非法 id 返回 `400`；把 `FindByID`/`Update`/`Delete` 三个 repository 方法接上调用方（backlog B3）。

## 可选挑战题

- 分页（`page`/`page_size`，破坏性改动 `GET /todos` 形状）。
- 状态筛选（`?done=true/false`，区分「不传」与「done=false」）。

## 答疑记录

- 教案阶段没有提问，直接确认可以开始实施。

## Quiz

**Q1. PUT `/todos/42`（42 不存在）如果直接用 `repo.Update` 里的 `db.Save`，会发生什么？我们的 `UpdateTodo` 靠哪一步挡住了这个行为？**

回答：会直接插入一个 42 的记录。先查 42 这条存不存在。

结果：正确。

标准答案：`Save` 对主键不存在的记录执行 INSERT（把 id=42 一并插入），造出一条 ghost 记录而不是报错。`UpdateTodo` 先 `FindByID(id)`，返回 `nil` 时直接 `NotFound`，`Save` 走不到 upsert 分支。

**Q2. `DeleteTodo`（选择 B）靠 `RowsAffected == 0` 判 404。删除一条已被软删除的记录也返回 404——为什么 `RowsAffected` 这时是 0？**

回答：因为被删除了，找不到了。

结果：方向正确，机制可补充。

标准答案：GORM 对带 `gorm.DeletedAt` 的模型，所有查询和软删除都自动加 `deleted_at IS NULL` 过滤。第一次删除把 `deleted_at` 打了时间戳；第二次 `Delete` 的 `UPDATE ... WHERE id=? AND deleted_at IS NULL` 匹配 0 行，`RowsAffected == 0` → 404。

**Q3. `:id` 的解析为什么放在 handler 层而不是 service 层？非法 id 返回 400 还是 500？**

回答：service 层只关注业务逻辑，参数处理放在 handler 层合适。返回 400，属于非法传参，不是服务内部错误。

结果：正确。

标准答案：与回答一致——「id 怎么从 URL 表达」是 HTTP 细节，service 只接收干净的 `uint`；非法 id 是客户端输入问题，走 `apperror.Validation`（400）。

**Q4. `UpdateTodoRequest.Done` 为什么不能加 `binding:"required"`？**

回答：因为如果加了，如果 done 是 false 会被拒掉直接报错。

结果：正确。

标准答案：`bool` 加 `required` 时 validator 把零值 `false` 当成「未提供」，合法的 `{"done":false}` 会被判 400。`Done` 允许为 `false`，所以不加 `required`，缺省即零值 `false`。

## 今日产出

- `internal/service/todo.go`：`TodoRepository` 接口加 `FindByID`/`Update`/`Delete`；新增 `GetTodo`/`UpdateTodo`/`DeleteTodo`。
- `internal/handler/todo.go`：`TodoService` 接口加三个方法；新增 `parseID`/`GetTodo`/`UpdateTodo`/`DeleteTodo` 与 `UpdateTodoRequest`。
- `internal/repository/todo.go`：`Delete` 签名改为 `(int64, error)` 返回 `RowsAffected`（选择 B）。
- `cmd/server/main.go`：注册 `GET`/`PUT`/`DELETE /todos/:id` 三条路由。
- 测试：`internal/service/todo_test.go`、`internal/handler/todo_test.go` 补齐三个接口的用例（含非法 id 400、404、`done:false` 不被拒、PUT 不 upsert）；`internal/repository/todo_test.go` 适配新签名。
- 文档：`docs/api/todo-api.md` 三个接口挪到「已实现」。

## 运行过的命令

```bash
go build ./...
go vet ./...
gofmt -l .
go test -count=1 ./internal/...
RUN_INTEGRATION_TESTS=true go test -count=1 ./internal/repository/
# 端到端：docker 起 MySQL + 起服务，curl 实测 GET/PUT/DELETE 的 200/400/404
```

## 遇到的问题

- **选择 A/B 混用**：初版 `Delete` 改成返回 `(int64, error)`（选择 B）却仍保留 `FindByID` 存在性检查、丢弃 `RowsAffected`，两头成本都付了。收敛到选择 B：`DeleteTodo` 去掉 `FindByID`，靠 `RowsAffected == 0` 判 404，单次查询。
- **DELETE 成功 body 形状**：教案原写 `{"data":null,"message":"ok"}`，实际因 `response.Body.Data` 带 `omitempty`，`Data: nil` 被省掉，真实 body 是 `{"message":"ok"}`。测试和教案、API 文档已同步修正。

## 关键收获

1. GORM `Save` 对主键不存在的记录会 upsert（INSERT），PUT 语义必须先判存在性，否则会造出 ghost 记录。
2. 软删除下 `RowsAffected` 仍能正确判 not-found：GORM 自动加 `deleted_at IS NULL`，重复删除匹配 0 行。
3. `omitempty` 会让 `Data: nil` 从成功信封里消失——响应形状要以实际序列化结果为准，不能想当然。
4. 分层边界：路径参数解析属于 handler，业务判断属于 service；非法输入返回 400 而非 500。

## 明日计划

- Day 25：JWT 鉴权（`login` 发 token，TODO 接口按 `user_id` 隔离，收紧 CORS 白名单、`AuthPlaceholder` 移到受保护路由组）。
