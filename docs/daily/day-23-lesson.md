# Day 23 教案：用户注册与登录

## 学习目标

学完今天，需要能够做到：

1. 创建 `internal/model/user.go`，落地 Day 22 已经定案的 `User` 表结构。
2. 用 `bcrypt` 做密码哈希，说清为什么不能自己发明哈希方案、也不能用 `md5`/`sha256` 直接哈希密码。
3. 实现 `POST /api/v1/users/register`：邮箱唯一性校验、密码哈希、返回不含密码的用户信息。
4. 实现 `POST /api/v1/users/login`：校验邮箱和密码，失败时不区分「邮箱不存在」和「密码错」。
5. 让 `internal/service`、`internal/handler` 的边界规则对 `user` 这条新链路依然成立：接口由使用方声明、错误走 `apperror` 统一出口。
6. 处理数据库层面的唯一性冲突（`uniqueIndex` 触发的 MySQL 错误），不让它的原始文本泄漏给客户端。

今天**不实现 JWT**。登录成功后暂时不发 token，Day 25 再接上。

---

## Day 23 的位置

Day 22 已经把决策做完了：`User` 表结构定了、`Todo` 保留 `Done bool`、schema 权威来源是 `AutoMigrate`。今天是 Week 4 的第二天，第一次真正给项目加一条新的业务链路（之前三周都只有 TODO 一条）。

这条链路会完整复用 Day 20 定下的三层架构和 Day 21 定下的错误处理规范——今天没有新的架构决策，是把已有的模式套到 `user` 上。如果发现某个模式套不上，那才是今天需要新决策的地方。

---

## 第一件事：`User` 模型

Day 22 已经在 `docs/architecture/todo-api.md` 里定案了字段：

```go
// internal/model/user.go
package model

import "time"

// User represents a registered account.
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"size:255;not null;uniqueIndex" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
```

两个字段值得复述一下原因，不只是抄定案：

- `Email` 的 `uniqueIndex` 是数据库约束，不是应用层查重的替代品，是它的**兜底**。应用层查重（`FindByEmail` 之后判断是否存在）在两个并发注册请求同时查到「不存在」时会一起插入成功，除非数据库有唯一索引把第二次插入拒掉。今天两层都要写：应用层查重给出更友好的错误信息，数据库约束保证并发下不会有两条一样的邮箱。
- `PasswordHash` 的 `json:"-"` 和 `Todo.DeletedAt` 是同一个模式：这个字段永远不该出现在任何响应里，用 tag 在序列化层就杜绝，而不是指望每个写返回值的地方都记得手动排除它。

`cmd/server/main.go` 里 `AutoMigrate` 要加上这个模型：

```go
if err := db.AutoMigrate(&model.Todo{}, &model.User{}); err != nil {
	log.Fatal(err)
}
```

---

## 第二件事：为什么是 `bcrypt`

### 不能用 `md5` / `sha256` 直接哈希密码

`md5("password123")` 对所有用户永远算出同一个值。攻击者只需要一份「常见密码 → 哈希值」的对照表（rainbow table），拿到数据库泄漏的哈希后直接查表，不需要暴力破解。`sha256` 本质上是同一个问题：它们都是为「验证数据完整性」设计的**快速**哈希——计算越快，被暴力破解和查表攻击的成本就越低，这正好和密码哈希的需求反着来。

### 不能自己发明方案

「加个固定 salt 再 sha256」比裸 hash 好一点，但固定 salt 本身会被泄漏（它就在代码里），攻击者只需要为这一个应用重新生成一次对照表。自己攒的方案几乎总会漏掉一些成熟方案已经踩过的坑（salt 怎么生成、怎么存、迭代次数怎么定），密码哈希是「不要自己发明」的经典场景。

### `bcrypt` 做对了什么

实测（`golang.org/x/crypto/bcrypt`，已经在 `go.sum` 里作为间接依赖存在，今天用上后会变成直接依赖）：

```go
hash, _ := bcrypt.GenerateFromPassword([]byte("hunter2"), bcrypt.DefaultCost)
// hash = "$2a$10$swALCRKMxGVFrxELraapBOQmDkqSO8iKksAeB0n0BaaNuP.287Yv."
// len(hash) = 60

bcrypt.CompareHashAndPassword(hash, []byte("hunter2"))  // nil，匹配
bcrypt.CompareHashAndPassword(hash, []byte("wrong"))     // crypto/bcrypt: hashedPassword is not the hash of the given password
```

几个关键点：

- **每次哈希同一个密码，结果都不一样**——`bcrypt.GenerateFromPassword` 内部自动生成随机 salt 并把它编码进返回值里（`$2a$10$` 后面那段就是 salt），不需要你自己管理 salt 的生成和存储。
- **返回值本身就是要存进数据库的完整字符串**，固定 60 字节，`PasswordHash string` 用 `size:255` 绰绰有余。
- **`bcrypt.DefaultCost` 控制计算成本**（当前是 10，意味着 `2^10` 轮迭代），成本越高越慢，越难被暴力破解——这是有意为之的慢，不是性能问题。
- **`CompareHashAndPassword` 不是"解密再比较"**，`bcrypt` 是单向哈希，无法从 hash 还原出密码；比较时是把候选密码用同样的 salt 和 cost 重新走一遍哈希流程，再比较结果。
- 实测还发现一个边界：`bcrypt` 对输入密码长度有硬限制。

  ```go
  long := make([]byte, 80) // 80 bytes
  bcrypt.GenerateFromPassword(long, bcrypt.DefaultCost)
  // 返回错误: bcrypt: password length exceeds 72 bytes
  ```

  超过 72 字节会直接返回错误，不是静默截断。今天写 `Register` 时这个错误要走 `apperror.Validation`，不能让它变成一个 500。

### 到底往哪一层放

`bcrypt.GenerateFromPassword` / `CompareHashAndPassword` 属于业务规则（「密码必须以哈希形式存储」是业务约束，不是数据访问细节，也不是 HTTP 细节），放在 **service 层**，和 Day 20 定的边界一致：service 不 import `gin` 也不 import `gorm`，但可以 import `golang.org/x/crypto/bcrypt`。

---

## 第三件事：`Register`

### service 层

```go
// internal/service/user.go
type UserRepository interface {
	Create(user *model.User) error
	FindByEmail(email string) (*model.User, error)
}

func (s *UserService) Register(email, password string) (*model.User, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, apperror.Validation("email is required")
	}
	if len(password) == 0 {
		return nil, apperror.Validation("password is required")
	}

	existing, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, apperror.Internal("register failed")
	}
	if existing != nil {
		return nil, apperror.Validation("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		// This is bcrypt's 72-byte input limit, not a server fault.
		return nil, apperror.Validation("password is too long")
	}

	user := &model.User{Email: email, PasswordHash: string(hash)}
	if err := s.repo.Create(user); err != nil {
		return nil, apperror.Internal("register failed")
	}
	return user, nil
}
```

`FindByEmail` 查重是应用层的第一道防线，给出「email already registered」这个明确的错误信息。但两个并发请求可能都通过这一步查重（都查到 `existing == nil`），然后都执行 `Create`——这时候数据库的 `uniqueIndex` 会拒掉第二个 `Create`，`repo.Create` 返回一个 MySQL 的重复键错误。

### repository 层要处理的唯一性冲突

实测这个错误的真实形状（临时建表验证，跑完即删）：

```text
Error 1062 (23000): Duplicate entry 'a@example.com' for key 'tmp_users.idx_tmp_users_email'
error type: *mysql.MySQLError
```

这段原文里的表名和索引名，和 Day 21 的绑定错误原文是同一类问题——不能直接冒泡给 handler 再冒泡给客户端。

### `TranslateError` 是前提条件，不是自动生效的

实测发现一个容易踩空的点：GORM 提供 `errors.Is(err, gorm.ErrDuplicatedKey)` 这个判断方式，但它**默认不生效**。验证过程：

```go
db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})          // 项目现状
// db.Create 遇到重复键时：
// err = "Error 1062 (23000): Duplicate entry 'dup@example.com' for key '...'"
// 类型是 *mysql.MySQLError
// errors.Is(err, gorm.ErrDuplicatedKey) == false

db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{TranslateError: true})
// 同样的重复键插入：
// err = "duplicated key not allowed"
// 类型是 *errors.errorString（即 gorm.ErrDuplicatedKey 本身）
// errors.Is(err, gorm.ErrDuplicatedKey) == true
```

`internal/config/gorm.go` 现在的 `ConnectGorm` 用的是裸 `&gorm.Config{}`，也就是说**如果今天不改这一行，`errors.Is(err, gorm.ErrDuplicatedKey)` 永远是 `false`**——这不是一个可以先跳过、以后再补的细节，是今天这条链路能不能工作的前提。今天要把它改成：

```go
// internal/config/gorm.go
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{TranslateError: true})
```

这一行改动同时影响 `TodoRepository` 的错误路径——`gorm.io/driver/mysql` 的翻译表目前只翻译 1062（重复键）这一种错误码，其他 GORM 语义错误（如 `ErrRecordNotFound`）本来就不经过这个翻译表，`Todo` 那边的行为不会变。

### repository 层的处理方式

```go
// internal/repository/user.go
func (r *UserRepository) Create(user *model.User) error {
	if err := r.db.Create(user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return apperror.Validation("email already registered")
		}
		return err
	}
	return nil
}
```

开启 `TranslateError` 后就不需要自己解析 MySQL 的错误码或错误文本——这是比手写字符串匹配更稳的写法，跨数据库驱动也一样成立（换成 Postgres 时判断逻辑不用改）。

这里出现了一个值得注意的分层问题：**repository 层直接返回了 `apperror.Validation`**，看起来违反了「repository 不该知道 HTTP 语义」的直觉。但 `apperror.Error` 本身不含任何 HTTP 细节（它是 `Code`/`Message`/`StatusCode` 三个字段的值类型，`StatusCode` 只是一个整数，repository 不需要 import `net/http` 或 `gin` 就能构造它）。真正的边界规则是「不 import `gin`、不 import `net/http` 处理请求」，`apperror` 是一个所有层都能共享的错误值类型，这一点和 Day 20 的边界规则并不冲突。

（如果你觉得 repository 直接返回业务语义的错误不舒服，另一种写法是 repository 只返回 `gorm.ErrDuplicatedKey` 本身，让 service 层用 `errors.Is` 去判断并翻译成 `apperror.Validation`——两种写法都成立，今天选哪种，把理由写进代码注释就行。）

### handler 层

```go
// internal/handler/user.go
type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("request_id=%s bind_error=%v", middleware.RequestIDFromContext(c), err)
		writeError(c, apperror.Validation("invalid request body"))
		return
	}

	user, err := h.service.Register(req.Email, req.Password)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Body{
		Data:    UserResponse{ID: user.ID, Email: user.Email},
		Message: "ok",
	})
}
```

和 `todo.go` 的 `CreateTodo` 是同一个模式：绑定失败记日志返回固定文案，业务错误统一走 `writeError`。`UserResponse` 只暴露 `ID`、`Email`，不会不小心带出 `PasswordHash`——即使模型的 `json:"-"` 已经挡了一层，handler 层单独定义响应结构体是第二层保险，两层独立生效。

`writeError` 目前在 `internal/handler/todo.go` 里是包级私有函数，`user.go` 在同一个包内可以直接复用，不需要改动。

---

## 第四件事：`Login`

```go
func (s *UserService) Login(email, password string) (*model.User, error) {
	email = strings.TrimSpace(email)
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, apperror.Internal("login failed")
	}
	if user == nil {
		return nil, apperror.Validation("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, apperror.Validation("invalid email or password")
	}

	return user, nil
}
```

关键点：**邮箱不存在**和**密码错误**返回**同一句话**、同一个错误码。如果分开返回「用户不存在」和「密码错误」，攻击者可以拿一份邮箱字典跑一遍注册状态的接口，筛出哪些邮箱已经注册过——这是一个真实的信息泄漏，属于 Day 21 学过的「面向客户端的信息 vs 面向自己的信息」在认证场景下的具体应用：区分两种失败原因对合法用户没有额外价值（反正都要重新输入），但对攻击者是有价值的账号枚举信息。

今天没有 JWT，`Login` 成功后 handler 暂时只返回 `{id, email}`，不发 token：

```go
c.JSON(http.StatusOK, response.Body{
	Data:    UserResponse{ID: user.ID, Email: user.Email},
	Message: "ok",
})
```

`docs/api/todo-api.md` 里「计划中」小节已经写了登录响应形状是 `{ "data": { "token" }, "message": "ok" }`，今天先不实现这部分，等 Day 25 JWT 落地时再把 `token` 字段加上——现在硬塞一个假 token 只会制造「看起来已经支持认证」的错觉。

---

## 第五件事：路由

```go
v1 := router.Group("/api/v1")
{
	v1.GET("/health", handler.HealthHandler)
	v1.POST("/todos", todoHandler.CreateTodo)
	v1.GET("/todos", todoHandler.ListTodos)
	v1.POST("/users/register", userHandler.Register)
	v1.POST("/users/login", userHandler.Login)
}
```

不新建路由组、不加中间件保护——这两个接口本来就应该公开访问（还没登录的人才需要注册/登录）。`AuthPlaceholder` 继续保持全局挂载、继续是无操作透传，这条不受今天改动影响。

---

## 今日代码结构

新增：

```
internal/model/user.go          # User 模型
internal/service/user.go        # Register / Login，含 bcrypt 调用
internal/repository/user.go     # Create / FindByEmail，处理重复键错误
internal/handler/user.go        # RegisterRequest / LoginRequest / UserHandler
```

改动：

```
cmd/server/main.go               # AutoMigrate 加 &model.User{}，注册两个路由，装配 UserRepository/UserService/UserHandler
internal/config/gorm.go          # gorm.Config 加 TranslateError: true，否则 errors.Is(err, gorm.ErrDuplicatedKey) 永远是 false
```

加测试（和 `todo_test.go` / `internal/service/todo_test.go` 同样的模式：service 用内存假 repository，handler 用假 service）：

```
internal/service/user_test.go
internal/handler/user_test.go
internal/repository/user_test.go   # RUN_INTEGRATION_TESTS 门控，和 todo_test.go 共用 TestMain
```

`internal/repository/user_test.go` 不需要自己的 `TestMain`——它和 `todo_test.go` 在同一个包里，`TestMain` 只能有一个，今天复用已有的那个，把 `AutoMigrate` 调用加上 `&model.User{}`。

---

## 今日验收标准

1. `POST /api/v1/users/register` 传合法邮箱密码，返回 `201` + `{id, email}`，不含密码。
2. 重复邮箱注册返回 `400`，错误信息不含表名、索引名等数据库细节。
3. `POST /api/v1/users/login` 密码正确返回 `200` + 用户信息；密码错误和邮箱不存在返回**完全相同**的错误信息和状态码。
4. `PasswordHash` 在任何响应里都不出现——包括直接 `curl` 观察响应体。
5. `internal/service`、`internal/handler`、`internal/repository` 三层边界规则不被打破（service/handler 不 import `gorm` 的规则对 `gorm.ErrDuplicatedKey` 这种情况要重新确认一下：`errors.Is` 判断放在 repository 层，service 拿到的已经是 `apperror.Error`，不需要知道 GORM 的存在）。
6. `internal/config/gorm.go` 的 `ConnectGorm` 已经开启 `TranslateError: true`，且用真实的并发/重复插入验证过 `errors.Is(err, gorm.ErrDuplicatedKey)` 确实生效（不能只看代码编译通过）。
7. `make fmt test vet` 全部通过。
8. `docs/api/todo-api.md` 补上 `register`/`login` 的「已实现」条目（原来在「计划中」小节，今天要挪过去并核对真实字段）。

## 可选挑战题

1. 邮箱唯一性和结构化错误响应（Issue 原文的可选挑战）：给 `RegisterRequest` 加邮箱格式校验（`binding:"required,email"`），实测一下格式非法时 Gin validator 的原始错误文本长什么样，确认它也需要走 Day 21 的「原文进日志、对外固定文案」这条规则。
2. 给 `internal/repository/user_test.go` 补一个真实触发数据库唯一索引冲突的用例（同一个邮箱 `Create` 两次），断言第二次返回的是 `apperror.Validation`，不是原始的 `*mysql.MySQLError`。

## 今天最容易踩的坑

1. **用 `md5`/`sha256` 哈希密码。** 这是这条业务线最容易出的错，因为标准库里 `crypto/md5`、`crypto/sha256` 触手可及，而 `bcrypt` 需要额外 import 一个子仓库包。手快的时候容易图方便用标准库里"看起来能用"的哈希函数。
2. **`Login` 对"邮箱不存在"和"密码错误"给出不同的错误信息。** 这是账号枚举攻击的入口，今天的验收标准第 3 条专门测这个。
3. **忘记数据库层的唯一索引兜底，只做应用层查重。** 应用层查重在并发请求下有竞态窗口，今天要两层都写，且实测验证过重复键错误的真实文本会泄漏表名和索引名。
4. **`UserResponse` 直接用 `model.User`，指望 `json:"-"` 兜底。** `json:"-"` 确实能挡住序列化，但 handler 层单独定义响应结构体是更明确的第二层保险——如果哪天 `PasswordHash` 的 tag 被不小心改掉，独立的响应结构体不会受影响。
5. **`bcrypt.GenerateFromPassword` 的错误被当成 500。** 72 字节长度限制会触发这个错误，它是客户端输入问题，应该走 `apperror.Validation`，不是 `apperror.Internal`。
6. **写了 `errors.Is(err, gorm.ErrDuplicatedKey)` 却忘记开 `TranslateError: true`。** 实测验证过：不开这个选项，`db.Create` 返回的永远是原始的 `*mysql.MySQLError`，`errors.Is` 永远判 `false`，重复邮箱注册会被误判成 `apperror.Internal`（500），而不是预期的 `apperror.Validation`（400）。这一行配置改动今天必须跟着 `user.go` 一起改，否则验收标准第 2 条测不过。
