# TODO API 文档

> 本文档分为「已实现」和「计划中」两部分。
>
> - **已实现**：当前 `main` 分支上真实跑得通的接口，示例响应都是从运行中的服务上抓下来的。改代码时必须同步改这一节。
> - **计划中**：设计意图，尚未落地。不要照着它写客户端。
>
> 最后核对：2026-09-22（Day 25，逐条 curl 验证）

## 基础信息

- Base URL：`http://localhost:8080`
- API Prefix：`/api/v1`
- 依赖：MySQL（启动方式见 README 的「本地启动」）
- 认证：**JWT（Day 25 起）**。`login` 成功返回 `token`（HS256，默认 24h 过期）；`/todos` 全部接口需要带 `Authorization: Bearer <token>`，缺失/非法/过期一律 `401`。公开接口：`GET /health`、`POST /users/register`、`POST /users/login`。签名密钥来自 `JWT_SECRET` 环境变量（未设置时用 dev 兜底，仅限本地）。**尚未按 `user_id` 隔离数据**——所有登录用户共享同一份 todos（`Todo.UserID` 留后续）。

## 全局约定

### 响应信封

成功：

```json
{ "data": {}, "message": "ok" }
```

失败：

```json
{ "error": { "code": "NOT_FOUND", "message": "route not found" } }
```

例外：`GET /api/v1/health` 目前返回裸对象，不走信封（见 `docs/issues-backlog.md` A7，待决策）。

### 全局响应头

所有响应（包括 404）都会带上：

| Header | 说明 |
|--------|------|
| `X-Request-Id` | 请求 ID。客户端可以自己传 `X-Request-ID`，服务端原样保留；不传则生成一个 32 位十六进制串。排查问题时把它给到服务端就能定位日志行 |
| `Access-Control-Allow-Origin` | 仅当请求 `Origin` 在白名单（`CORS_ALLOWED_ORIGINS`，默认 `http://localhost:3000`）内时回显该 Origin；否则不设置该头（`docs/issues-backlog.md` A5，Day 25 收紧） |
| `Access-Control-Allow-Methods` | `GET,POST,PUT,PATCH,DELETE,OPTIONS` |
| `Access-Control-Allow-Headers` | `Origin, Content-Type, Accept, Authorization, X-Request-ID` |
| `Access-Control-Expose-Headers` | `X-Request-ID` |
| `Vary` | `Origin` |

`OPTIONS` 预检请求返回 `204 No Content`，不进入业务 handler。

### 错误响应不回显内部信息

请求体校验失败时，服务端返回固定文案，validator/json 的原始错误只写日志（按 `request_id` 关联），不返回给客户端。

---

# 一、已实现

> **认证边界**：下面的 `/todos` 四个接口都需要 `Authorization: Bearer <token>`；`GET /health`、`POST /users/register`、`POST /users/login` 公开。
>
> 受保护接口在缺失/格式错误的 Authorization 头时返回 `401`：
>
> ```json
> { "error": { "code": "UNAUTHORIZED", "message": "missing or malformed authorization header" } }
> ```
>
> token 非法/过期/签名错时返回 `401`（原因只写日志，客户端只收笼统文案）：
>
> ```json
> { "error": { "code": "UNAUTHORIZED", "message": "invalid or expired token" } }
> ```

## GET /api/v1/health

认证：不需要。

响应 `200`：

```json
{
  "status": "ok",
  "service": "go-learning",
  "version": "0.1.0"
}
```

注意：路径是 `/api/v1/health`，**不是** `/health`。`/health` 会命中 404。

## POST /api/v1/todos

请求体：

```json
{ "title": "learn go" }
```

`title` 必填。当前没有其他字段（`description`、`due_date` 属于计划中）。

响应 `201`：

```json
{
  "data": { "id": 3, "title": "learn go", "done": false },
  "message": "ok"
}
```

失败 `400`：

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "invalid request body: title is required"
  }
}
```

`title` 缺失、JSON 截断、JSON 类型不对（例如传数组）都返回这同一个 400 文案，细节在服务端日志里。

`title` 只有空白字符时也是 400：service 层会先 `TrimSpace`。

## GET /api/v1/todos

当前不支持分页和筛选参数，一次返回全部。

响应 `200`：

```json
{
  "data": [
    { "id": 1, "title": "learn go", "done": false },
    { "id": 2, "title": "write tests", "done": false }
  ],
  "message": "ok"
}
```

列表为空时返回 `"data": []`，不是 `null`。

## GET /api/v1/todos/{id}

按 ID 查询单个 todo。

响应 `200`：

```json
{
  "data": { "id": 1, "title": "learn go", "done": false },
  "message": "ok"
}
```

不存在返回 `404`：

```json
{ "error": { "code": "NOT_FOUND", "message": "todo not found" } }
```

`id` 不是合法的无符号整数（如 `/todos/abc`、负数、溢出）返回 `400`：

```json
{ "error": { "code": "VALIDATION_ERROR", "message": "invalid id" } }
```

## PUT /api/v1/todos/{id}

全量更新一个 todo 的 `title` 和 `done`。

请求体：

```json
{ "title": "updated title", "done": true }
```

`title` 必填；`done` 可选，缺省为 `false`（`done` 不加 `required`——布尔字段的 `required` 会把合法的 `false` 误判为「未提供」）。

响应 `200`：

```json
{
  "data": { "id": 1, "title": "updated title", "done": true },
  "message": "ok"
}
```

不存在返回 `404`（**不会 upsert 出一条新记录**）：

```json
{ "error": { "code": "NOT_FOUND", "message": "todo not found" } }
```

`title` 为空或只有空白返回 `400`（`{"code":"VALIDATION_ERROR","message":"title is required"}`）；`title` 缺失或 JSON 非法返回 `400`（`invalid request body`）；非法 `id` 返回 `400`（`invalid id`）。

## DELETE /api/v1/todos/{id}

删除一个 todo（模型是软删除，`deleted_at` 打标记，后续查询自动过滤）。

响应 `200`：

```json
{ "message": "ok" }
```

注意成功响应**没有 `data` 字段**：`data` 用了 `omitempty`，删除没有返回体，`Data` 为 `nil` 时被省掉。

不存在返回 `404`（`{"code":"NOT_FOUND","message":"todo not found"}`）；非法 `id` 返回 `400`（`invalid id`）。删除后再 `GET /todos/{id}` 返回 `404`，列表里也不再出现。

## 未匹配的路由

任何未注册的路径（任意 method）响应 `404`：

```json
{ "error": { "code": "NOT_FOUND", "message": "route not found" } }
```

响应里不回显请求的路径。404 同样带上全局响应头。

## 服务端 panic

handler 内 panic 时由 `Recovery` 中间件兜住，响应 `500`：

```json
{ "error": { "code": "INTERNAL_ERROR", "message": "internal server error" } }
```

panic 值和调用栈只写日志，按 `request_id` 关联。

## POST /api/v1/users/register

认证：不需要。

请求体：

```json
{ "email": "a@example.com", "password": "hunter2" }
```

`email`、`password` 均必填。密码用 `bcrypt` 哈希后存储，`PasswordHash` 永不出现在任何响应里。

响应 `201`：

```json
{
  "data": { "id": 1, "email": "a@example.com" },
  "message": "ok"
}
```

失败 `400`（邮箱已注册）：

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "email already registered"
  }
}
```

邮箱唯一性有两层保证：应用层先查一次（给出上面这个错误信息），数据库的 `uniqueIndex` 兜底并发场景下的竞态。两层触发的错误信息相同，客户端无法区分是哪一层拦下的。

密码超过 72 字节（`bcrypt` 的硬限制）也返回 `400`：

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "password is too long"
  }
}
```

## POST /api/v1/users/login

认证：不需要。

请求体：

```json
{ "email": "a@example.com", "password": "hunter2" }
```

响应 `200`：

```json
{
  "data": {
    "id": 1,
    "email": "a@example.com",
    "token": "eyJhbGciOiJIUzI1Ni...<省略>...sJ9.abc123"
  },
  "message": "ok"
}
```

`token` 是 HS256 签名的 JWT，默认 24h 过期，`sub` 是 userID。带上 `Authorization: Bearer <token>` 访问 `/todos` 即可。

失败 `400`：

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "invalid email or password"
  }
}
```

**邮箱不存在和密码错误返回完全相同的错误**（同一个 `message`、同一个状态码），刻意如此：区分两者会让攻击者可以用一份邮箱字典探测出哪些邮箱在系统里注册过。

---

# 二、计划中

以下接口**尚未实现**，字段和形状都可能变。

## 用户接口鉴权（Day 25）

`register`/`login` 已在 Day 23 实现，JWT 鉴权已在 Day 25 落地（见「已实现」：`login` 返回 `token`，`/todos` 需 Bearer）。**仍未做**的是按 `user_id` 隔离数据——当前所有登录用户共享同一份 todos，`Todo.UserID` 留后续。届时 `RequireAuth` 已把 `user_id` 存进 context（`middleware.UserIDFromContext`），service/repository 按它过滤即可。

## TODO 接口补全（Day 22-24）

`GET /todos/{id}`、`PUT /todos/{id}`、`DELETE /todos/{id}` 已在 Day 24 实现（见上「已实现」）。

## 列表分页与筛选（Day 24）

计划的查询参数与响应形状：

| 参数 | 说明 |
|------|------|
| `page` | 页码，默认 1 |
| `page_size` | 每页数量，默认 20 |
| `status` | 可选筛选条件 |

```json
{
  "data": { "items": [], "page": 1, "page_size": 20, "total": 0 },
  "message": "ok"
}
```

注意这与当前 `GET /api/v1/todos` 直接返回数组的形状**不兼容**，切换时是一次破坏性改动。

## 待决策的字段

`description`、`due_date`，以及状态用 `done bool` 还是 `status string`（见 `docs/issues-backlog.md` D3，Day 22 决策）。

---

## 错误码

| Code | HTTP | 状态 |
|------|------|------|
| `VALIDATION_ERROR` | 400 | 已实现 |
| `NOT_FOUND` | 404 | 已实现 |
| `INTERNAL_ERROR` | 500 | 已实现 |
| `UNAUTHORIZED` | 401 | 已实现（Day 25） |
| `FORBIDDEN` | 403 | 计划中（按 user_id 隔离后） |

