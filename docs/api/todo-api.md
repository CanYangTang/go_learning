# TODO API 文档

> 本文档分为「已实现」和「计划中」两部分。
>
> - **已实现**：当前 `main` 分支上真实跑得通的接口，示例响应都是从运行中的服务上抓下来的。改代码时必须同步改这一节。
> - **计划中**：设计意图，尚未落地。不要照着它写客户端。
>
> 最后核对：2026-09-11（Day 23，逐条 curl 验证）

## 基础信息

- Base URL：`http://localhost:8080`
- API Prefix：`/api/v1`
- 依赖：MySQL（启动方式见 README 的「本地启动」）
- 认证：**当前没有**。`AuthPlaceholder` 中间件已挂在全局，但不做任何校验，所有接口都是公开的。JWT 在 Day 25 实现，`register`/`login` 现在也不发 token。

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
| `Access-Control-Allow-Origin` | 当前**回显请求的 `Origin`**，没有白名单（`docs/issues-backlog.md` A5） |
| `Access-Control-Allow-Methods` | `GET,POST,PUT,PATCH,DELETE,OPTIONS` |
| `Access-Control-Allow-Headers` | `Origin, Content-Type, Accept, Authorization, X-Request-ID` |
| `Access-Control-Expose-Headers` | `X-Request-ID` |
| `Vary` | `Origin` |

`OPTIONS` 预检请求返回 `204 No Content`，不进入业务 handler。

### 错误响应不回显内部信息

请求体校验失败时，服务端返回固定文案，validator/json 的原始错误只写日志（按 `request_id` 关联），不返回给客户端。

---

# 一、已实现

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
  "data": { "id": 1, "email": "a@example.com" },
  "message": "ok"
}
```

不发 token——Day 25 JWT 落地后才会在 `data` 里加 `token` 字段。

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

`register`/`login` 已在 Day 23 实现（见上）。Day 25 会在 `login` 成功响应里加 `token` 字段，TODO 接口需要带 `Authorization: Bearer <token>`，并按 `user_id` 隔离数据。

## TODO 接口补全（Day 22-24）

| 接口 | 说明 |
|------|------|
| `GET /api/v1/todos/{id}` | 按 ID 查询 |
| `PUT /api/v1/todos/{id}` | 全量更新 |
| `DELETE /api/v1/todos/{id}` | 删除（模型已支持软删除） |

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
| `UNAUTHORIZED` | 401 | 计划中（Day 25） |
| `FORBIDDEN` | 403 | 计划中（Day 25） |

