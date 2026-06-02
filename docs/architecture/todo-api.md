# TODO API 架构说明

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
- 初期可使用内存实现。
- 后续切换到 GORM + MySQL。

### model

- 定义 User、TODO 等核心数据模型。
- 后续增加 GORM tag 和 JSON tag。

### middleware

- 日志。
- CORS。
- JWT 认证。
- 请求 ID。

## 核心模型规划

### User

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | uint | 主键 |
| Email | string | 邮箱，唯一 |
| PasswordHash | string | 哈希后的密码 |
| CreatedAt | time.Time | 创建时间 |
| UpdatedAt | time.Time | 更新时间 |

### TODO

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

- 服务可以本地启动。
- 注册登录可用。
- JWT 能保护 TODO 路由。
- TODO 支持创建、查询、更新、删除。
- 列表支持分页和状态筛选。
- 数据持久化到 MySQL。
- `go test ./...` 通过。
- README 和 API 文档能指导别人运行项目。
