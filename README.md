# Go Learning

这是一个用 AI 全程管理的 Go 后端学习项目。目标是在 1 个月内通过持续代码产出，掌握 Go 后端开发基础，并完成一个增强版 TODO API。

## 学习方式

- 每天 1-2 小时，以代码产出为主。
- 每天都有核心任务和可选挑战题。
- 每个任务都需要有可验证产出。
- 使用 GitHub Issues 和 Milestones 跟踪进度。
- 使用 AI 作为项目经理、导师、结对开发者和代码 Reviewer。

## 路线图

| 周期 | 主题 | 目标 |
|------|------|------|
| Week 1 | Go Foundations | 掌握 Go 基础语法、函数、错误处理、包管理和测试基础 |
| Week 2 | HTTP, JSON, Concurrency | 掌握并发、channel、标准库 HTTP、JSON、文件操作 |
| Week 3 | Gin, GORM, Project Architecture | 掌握 Gin、GORM、MySQL、中间件和分层架构 |
| Week 4 | Complete TODO API | 完成带认证、持久化、测试、文档和部署准备的 TODO API |

## 最终项目：TODO API

增强版 TODO API 将包含：

- 用户注册和登录
- 密码哈希
- JWT 认证
- TODO CRUD
- 分页和筛选
- MySQL 持久化
- Gin + GORM
- 认证、日志、CORS 中间件
- 单元测试和基础集成测试
- API 文档
- Docker Compose 部署准备

## 项目结构

```text
.
├── cmd/server/              # 服务入口
├── internal/                # 后端内部实现
│   ├── config/
│   ├── handler/
│   ├── middleware/
│   ├── model/
│   ├── repository/
│   └── service/
├── pkg/                     # 可复用辅助包
│   ├── apperror/
│   └── response/
├── docs/
│   ├── api/                 # API 文档
│   ├── architecture/        # 架构说明
│   ├── daily/               # 每日学习记录
│   └── weekly/              # 每周复盘
├── deployments/             # 部署配置
├── scripts/                 # 辅助脚本
├── test/                    # 集成测试支持
├── GO_LEARN_PLAN.md         # 总学习计划
├── Makefile                 # 常用命令
└── go.mod
```

## 常用命令

```bash
make fmt
make test
make vet
make run
```

等价命令：

```bash
go fmt ./...
go test ./...
go vet ./...
go run ./cmd/server
```

## 每日 AI 工作流

1. 选择当天 GitHub Issue。
2. 让 AI 根据 Issue 给出当天的核心任务和验收标准。
3. 先独立实现核心任务。
4. 卡住时让 AI 给提示，而不是直接给完整答案。
5. 完成后运行格式化、测试和验证命令。
6. 让 AI Review 今日代码。
7. 更新 `docs/daily/day-NN.md`。
8. 提交当天成果。

## 跟踪文档

- 学习计划：`GO_LEARN_PLAN.md`
- 每日记录：`docs/daily/`
- 每周复盘：`docs/weekly/`
- 架构说明：`docs/architecture/todo-api.md`
- API 文档：`docs/api/todo-api.md`
