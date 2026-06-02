# TODO API 文档

## 基础信息

- Base URL：`http://localhost:8080`
- API Prefix：`/api/v1`
- Auth：JWT Bearer Token

## 健康检查

### GET /health

认证：不需要

响应：

```json
{
  "status": "ok",
  "service": "go-learning"
}
```

## 用户接口

### POST /api/v1/users/register

认证：不需要

请求体：

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

响应：

```json
{
  "data": {
    "id": 1,
    "email": "user@example.com"
  },
  "message": "ok"
}
```

### POST /api/v1/users/login

认证：不需要

请求体：

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

响应：

```json
{
  "data": {
    "token": "jwt-token"
  },
  "message": "ok"
}
```

## TODO 接口

以下接口均需要认证：

```text
Authorization: Bearer <token>
```

### GET /api/v1/todos

查询参数：

| 参数 | 说明 |
|------|------|
| page | 页码，默认 1 |
| page_size | 每页数量，默认 20 |
| status | 可选，pending 或 done |

响应：

```json
{
  "data": {
    "items": [],
    "page": 1,
    "page_size": 20,
    "total": 0
  },
  "message": "ok"
}
```

### POST /api/v1/todos

请求体：

```json
{
  "title": "learn go",
  "description": "finish today's core task",
  "due_date": "2026-06-30T00:00:00Z"
}
```

### GET /api/v1/todos/{id}

根据 ID 查询 TODO。

### PUT /api/v1/todos/{id}

请求体：

```json
{
  "title": "learn go backend",
  "description": "update task description",
  "status": "done",
  "due_date": "2026-06-30T00:00:00Z"
}
```

### DELETE /api/v1/todos/{id}

删除指定 TODO。

## 错误码规划

| Code | 说明 |
|------|------|
| VALIDATION_ERROR | 请求参数错误 |
| UNAUTHORIZED | 未认证或 token 无效 |
| FORBIDDEN | 无权限访问资源 |
| NOT_FOUND | 资源不存在 |
| INTERNAL_ERROR | 服务内部错误 |
