# Go 后端学习计划（AI 管理版）

> 目标：每天 1-2 小时，一个月具备 Go 后端开发基础能力，并完成一个增强版 TODO API。

---

## 学习原则

1. **代码优先**：看 30 分钟，写 60 分钟。
2. **每日可验证**：每天都有核心任务、可选挑战、产出和验证命令。
3. **项目串联**：所有练习逐步服务于最终 TODO API。
4. **AI 全程协作**：AI 作为项目经理、导师、结对开发者和 Reviewer。
5. **不抄代码**：先自己写，卡住时让 AI 给提示，最后再做 Review。

---

## 每日学习模板

```text
1. [10min] 回顾昨天内容和未解决问题
2. [30min] 学习今天核心概念
3. [60min] 完成核心任务代码
4. [可选] 完成挑战题
5. [10min] 运行验证命令
6. [10min] 记录今日总结
```

每日记录写入：`docs/daily/day-NN.md`。

---

## Week 1：Go Foundations

### 目标

掌握 Go 基础语法、函数、错误处理、包管理和测试基础，形成每天写 Go 代码的节奏。

| Day | 主题 | 核心任务 | 可选挑战题 | 产出 | 验证 |
|-----|------|----------|------------|------|------|
| Day 1 | 项目初始化 | 初始化 Git、Go module、README、目录结构和 `/health` 服务 | 给 `/health` 响应增加版本字段 | `go.mod`、`README.md`、`cmd/server/main.go` | `make run`、访问 `/health` |
| Day 2 | 变量和类型 | 编写变量、常量、基础类型和类型转换练习 | 总结 `int`、`float64`、`string` 转换常见坑 | `week01-basics/day02-syntax/` | `go test ./...` |
| Day 3 | 流程控制 | 实现九九乘法表和基础分支判断练习 | 实现一个简单命令行菜单 | `week01-basics/day03-control-flow/` | `go run` 或测试通过 |
| Day 4 | 函数 | 实现支持加减乘除的计算器函数 | 增加变参、匿名函数或闭包版本 | `week01-basics/day04-functions/` | `go test ./...` |
| Day 5 | 错误处理 | 实现显式错误返回、`defer` 和基础错误包装练习 | 对计算器增加错误类型 | `week01-basics/day05-errors/` | `go test ./...` |
| Day 6 | 包管理 | 拆分 package，练习 `go mod` 和包导入 | 编写一个可复用工具包 | `week01-basics/day06-packages/`、`pkg/` | `go list ./...` |
| Day 7 | 周总结 | 完成工具函数包和 Week 1 复盘 | 给工具包补测试和 README 说明 | `docs/weekly/week-01.md` | `make fmt test vet` |

### Week 1 检查点

- 能独立创建和运行 Go 程序。
- 理解变量、函数、流程控制和错误处理。
- 能使用 `go test`、`go fmt`、`go vet`。
- 每天都有代码和学习记录。

---

## Week 2：HTTP, JSON, Concurrency

### 目标

掌握 Go 并发模型和常用标准库，能使用 `net/http` 写基础 HTTP 服务。

| Day | 主题 | 核心任务 | 可选挑战题 | 产出 | 验证 |
|-----|------|----------|------------|------|------|
| Day 8 | Goroutine | 使用 goroutine 和 `sync.WaitGroup` 并发执行任务 | 比较串行和并发耗时 | `week02-core/day08-goroutine/` | `go test ./...` |
| Day 9 | Channel | 使用无缓冲、有缓冲 channel 实现协程通信 | 使用 `select` 增加超时控制 | `week02-core/day09-channel/` | `go test ./...` |
| Day 10 | 并发模式 | 实现生产者消费者任务队列 | 增加 worker 数量配置和优雅关闭 | `week02-core/day10-task-queue/` | `go test ./...` |
| Day 11 | HTTP 服务 | 用 `net/http` 完善 `/health` 和基础路由 | 增加简单请求日志 | `cmd/server/` | `make run`、curl 验证 |
| Day 12 | JSON 处理 | 实现 JSON 编码、解码和结构体 tag 练习 | 为 TODO 请求增加基础校验 | `week02-core/day12-json/`、`pkg/response/` | `go test ./...` |
| Day 13 | 文件操作 | 使用 `os`、`bufio`、`io` 完成文件读写工具 | 读取 JSON 文件并解析为结构体 | `week02-core/day13-files/` | `go test ./...` |
| Day 14 | 综合练习 | 完成 HTTP + JSON + 并发的综合练习 | 增加 context 超时取消 | `week02-core/day14-summary/` | `make fmt test vet` |

### Week 2 检查点

- 理解 goroutine、channel、select 和 WaitGroup。
- 能写基础 HTTP 服务和 JSON API。
- 能识别常见并发问题，例如 goroutine 泄漏和共享状态竞争。

---

## Week 3：Gin, GORM, Project Architecture

### 目标

掌握 Gin、GORM、MySQL 和分层架构，逐步把练习升级为真实 TODO API。

| Day | 主题 | 核心任务 | 可选挑战题 | 产出 | 验证 |
|-----|------|----------|------------|------|------|
| Day 15 | Gin 入门 | 引入 Gin，创建 `/api/v1` 路由组 | 增加统一 404 响应 | `cmd/server/`、`internal/handler/` | `make run`、curl 验证 |
| Day 16 | 请求响应 | 实现 TODO 创建和列表接口的请求绑定与 JSON 响应 | 增加请求参数校验 | `internal/handler/`、`pkg/response/` | `go test ./...` |
| Day 17 | MySQL 基础 | 增加 MySQL 连接配置和基础 CRUD 练习 | 使用 Docker Compose 启动 MySQL | `internal/config/`、`deployments/` | DB 连接成功 |
| Day 18 | GORM | 定义 GORM 模型并实现 TODO repository | 增加事务练习 | `internal/model/`、`internal/repository/` | `go test ./...` |
| Day 19 | 中间件 | 实现日志、认证占位、CORS 中间件 | 增加请求 ID | `internal/middleware/` | curl 验证响应头 |
| Day 20 | 分层架构 | 重构为 handler/service/repository 分层 | 为 service 增加单元测试 | `internal/service/` | `make fmt test vet` |
| Day 21 | 周总结 | 完成 API 服务雏形和 Week 3 复盘 | 补充架构图和接口文档 | `docs/weekly/week-03.md` | `make fmt test vet` |

### Week 3 检查点

- 能使用 Gin 构建 REST API。
- 理解 handler/service/repository 的职责边界。
- 能使用 GORM 操作 MySQL。
- TODO API 雏形可以运行。

---

## Week 4：Complete TODO API

### 目标

完成增强版 TODO API，具备认证、持久化、测试、文档和部署准备。

| Day | 主题 | 核心任务 | 可选挑战题 | 产出 | 验证 |
|-----|------|----------|------------|------|------|
| Day 22 | 数据库设计 | 完成 User、TODO 表设计和项目结构整理 | 补充 ER 图或表结构说明 | `docs/architecture/todo-api.md` | 文档完整 |
| Day 23 | 用户模块 | 实现注册、登录和密码哈希 | 增加邮箱唯一性和错误响应 | `internal/handler/`、`internal/service/`、`internal/model/` | 注册登录成功 |
| Day 24 | TODO CRUD | 实现创建、查询、更新、删除 TODO | 增加分页和状态筛选 | TODO API 核心接口 | curl 验证 CRUD |
| Day 25 | 认证和日志 | 实现 JWT 中间件，保护 TODO 路由 | 增加 token 过期处理 | `internal/middleware/auth.go` | 无 token 被拒绝 |
| Day 26 | 重构优化 | 统一错误处理、响应格式和分层边界 | 增加配置加载优化 | `pkg/apperror/`、`pkg/response/` | `make fmt test vet` |
| Day 27 | 测试和文档 | 增加单元测试、API 文档和使用示例 | 增加集成测试或 HTTP 示例 | `*_test.go`、`docs/api/todo-api.md` | `go test ./...` |
| Day 28 | 部署和总结 | 增加 Docker Compose 准备和最终复盘 | 写完整项目 README 使用说明 | `deployments/`、`docs/weekly/week-04.md` | 项目可按 README 启动 |

### Week 4 检查点

- TODO API 支持注册登录和 JWT 认证。
- TODO CRUD 支持持久化、分页和筛选。
- 项目有测试、文档和部署准备。
- 能向别人清楚介绍项目结构和关键技术取舍。

---

## AI 协作方式

### 每天开始

```text
根据当前 GitHub Issue 和昨天的进度，给我今天的核心任务、可选挑战和验收标准。
```

### 卡住时

```text
我卡在这里了。先给我提示，不要直接给完整答案。
```

### 完成后

```text
Review 我今天的代码，重点看 Go 风格、错误处理、测试和工程结构。
```

### 每周结束

```text
根据本周代码和 daily notes，帮我做一次周复盘，指出下周要补强的点。
```

---

## 推荐资源

| 资源 | 类型 | 说明 |
|------|------|------|
| https://go.dev/tour/ | 官方教程 | Go 官方入门教程 |
| https://pkg.go.dev/ | 官方文档 | 标准库和第三方库文档 |
| https://gobyexample.com/ | 示例 | 按主题学习 Go 示例 |
| https://gin-gonic.com/ | 框架文档 | Gin 官方文档 |
| https://gorm.io/ | ORM 文档 | GORM 官方文档 |

---

## 完成标准

一个任务只有同时满足以下条件才算完成：

- 核心代码已经实现。
- 必要测试或手动验证已经完成。
- `go fmt ./...` 通过。
- `go test ./...` 通过。
- 今日学习记录已经更新。
- 如果是项目功能，API 文档或 README 已同步更新。
