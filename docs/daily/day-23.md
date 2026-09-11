# Day 23 学习记录

## 日期

2026-09-10

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/23

## 今日教案

- 教案文档：`docs/daily/day-23-lesson.md`

## 核心任务

- 创建 `internal/model/user.go`，落地 Day 22 定案的 `User` 表结构。
- 用 `bcrypt` 做密码哈希，说清为什么不能用 `md5`/`sha256`、也不能自己发明方案。
- 实现 `POST /api/v1/users/register`：邮箱唯一性校验（应用层 + 数据库唯一索引兜底）、密码哈希、返回不含密码的用户信息。
- 实现 `POST /api/v1/users/login`：邮箱不存在和密码错误返回完全相同的错误信息，避免账号枚举。
- `internal/config/gorm.go` 开启 `TranslateError: true`，让 `errors.Is(err, gorm.ErrDuplicatedKey)` 生效。

## 答疑记录

- 教案阶段没有提问，直接确认可以开始实施。

## Quiz

**Q1. `Register` 里，`FindByEmail` 查重通过后 `Create` 仍然可能返回重复键错误——这是一个真实的并发竞态，而不是理论上的边界情况。请说说：如果没有 `Email` 的 `uniqueIndex`，只靠应用层的 `FindByEmail` 查重，两个并发请求同时到达会发生什么？**

回答：数据库里出现两条邮箱完全相同的用户记录。

结果：部分正确，漏了机制说明。

标准答案：两个请求的 `FindByEmail` 都在对方 `Create` 完成前执行，都查到 `nil`（判定邮箱不存在），于是都通过校验并各自插入成功，产生两条邮箱相同的记录——这是查重和插入之间的时间窗口（check-then-act 竞态），没有数据库约束就无法拒绝第二次插入。

**Q2. `internal/config/gorm.go` 里 `TranslateError: true` 具体做了什么？如果去掉这个配置，`errors.Is(err, gorm.ErrDuplicatedKey)` 会返回什么结果，`Create` 遇到重复邮箱时最终会给客户端返回什么状态码？**

回答：`TranslateError: true` 让 GORM 识别底层驱动错误码（如 MySQL 1062），翻译成 `gorm.ErrDuplicatedKey`。去掉后 `errors.Is` 判 `false`，客户端本该收到的 400 会变成 500——把用户输入问题误分类成服务器故障，对客户端和监控都是误导。

结果：正确。

标准答案：与回答一致。

**Q3. `Login` 为什么要让「邮箱不存在」和「密码错误」返回完全相同的错误？如果分别返回不同信息，攻击者能做什么？**

回答：为了更安全。攻击者不需要知道密码，只靠错误信息不同就能批量探测哪些邮箱已注册。

结果：方向正确，攻击链条可以展开得更具体。

标准答案：如果响应可区分，攻击者可以拿一份邮箱字典（如数据泄露库里的常见邮箱）逐个提交，通过响应差异筛出哪些邮箱在本系统注册过——这本身是隐私泄露（暴露某人是否是该网站用户），也是撞库、钓鱼攻击的前置信息收集步骤。

## 可选挑战题

-

## 今日产出

- `internal/model/user.go`：`User` 模型。
- `internal/service/user.go`：`Register`/`Login`，含 bcrypt 哈希、72 字节限制处理、并发重复键错误透传。
- `internal/repository/user.go`：`Create`/`FindByEmail`，`errors.Is(err, gorm.ErrDuplicatedKey)` 处理重复邮箱。
- `internal/handler/user.go`：`RegisterRequest`/`LoginRequest`/`UserHandler`。
- `internal/config/gorm.go`：`TranslateError: true`。
- `cmd/server/main.go`：装配 `User` 相关依赖，注册 `/register`、`/login` 路由。
- 测试：`internal/service/user_test.go`、`internal/handler/user_test.go`、`internal/repository/user_test.go`。

## 运行过的命令

```bash
go build ./...
go vet ./...
gofmt -l .
go mod tidy
go test -count=1 ./internal/...
```

## 遇到的问题

- 代码审查发现一个 bug：`Register` 里 `s.repo.Create` 失败时统一包成 `apperror.Internal`，丢掉了 repository 已经翻译好的 `apperror.Validation("email already registered")`——并发场景下重复邮箱注册会误判成 500。用 `errors.As` 透传已有的 `apperror.Error` 修复，测试 `TestRegisterPropagatesRepositoryValidationError` 验证。
- 三个新文件的 import 分组不符合项目惯例（标准库/第三方/本项目未分组），顺手理顺。

## 关键收获

1. 应用层查重和数据库唯一索引要两层都做：前者给出更友好的错误信息，后者堵住并发竞态窗口——两者缺一不可。
2. GORM 的 `TranslateError` 不是默认开启的，`errors.Is(err, gorm.ErrDuplicatedKey)` 依赖这个配置才能生效，否则错误分类会从「客户端问题」退化成「服务器故障」。
3. 认证场景下「邮箱不存在」和「密码错误」必须返回同一个错误，区分两者会变成账号枚举攻击的入口。

## 明日计划

- Day 24：TODO 完整 CRUD。

