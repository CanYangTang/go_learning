# Day 15 教案：Gin 框架和路由组

## 学习目标

学完今天，需要能够做到：

1. 理解 Gin 框架的基本用法。
2. 理解 Gin 的路由和路由组（route group）。
3. 使用 `gin.New()` 创建路由器。
4. 使用 `router.Group()` 创建路由组。
5. 实现 Gin handler 并返回 JSON 响应。
6. 编写 Gin handler 的测试。
7. 理解项目分层：`cmd/server` 和 `internal/handler` 的职责。

---

## Day 15 的位置

Week 2 学了 Go 标准库的 HTTP server、JSON、并发等基础知识。

Week 3 开始引入框架和架构分层：

- Day 15：Gin 框架和路由组
- Day 16：TODO 请求绑定和 JSON 响应
- Day 17：MySQL 连接和基础 CRUD
- Day 18：GORM 模型和 repository 实现
- Day 19：日志、认证、CORS 中间件
- Day 20：handler/service/repository 分层
- Day 21：API 服务骨架和周总结

今天的目标是用 Gin 替换标准库 `net/http`，并建立 `/api/v1` 路由组。

---

## Gin 是什么

Gin 是 Go 语言最流行的 Web 框架之一。

特点：

- 高性能（基于 HttpRouter）。
- 路由分组、中间件、请求绑定、JSON 响应等开箱即用。
- API 风格简洁，类似 Martini。

安装：

```bash
go get -u github.com/gin-gonic/gin
```

---

## Gin 基本用法

### 创建路由器

```go
router := gin.New()
```

`gin.New()` 创建一个空的 Gin 路由器。

也可以用 `gin.Default()`，它会自动附加 Logger 和 Recovery 中间件。

### 注册路由

```go
router.GET("/health", HealthHandler)
```

- `GET`：HTTP 方法。
- `"/health"`：路由路径。
- `HealthHandler`：处理函数。

### 启动服务

```go
router.Run(":8080")
```

`Run` 会启动 HTTP 服务器，监听指定端口。

---

## Gin Handler

Gin handler 的签名是：

```go
func(c *gin.Context)
```

`gin.Context` 封装了请求和响应：

```go
func HealthHandler(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "ok",
    })
}
```

- `c.JSON(code, obj)`：返回 JSON 响应，自动设置 `Content-Type: application/json`。
- `gin.H`：`map[string]any` 的简写，方便构造 JSON 对象。

---

## 路由组（Route Group）

路由组可以把一组路由放在同一个前缀下。

```go
v1 := router.Group("/api/v1")
{
    v1.GET("/health", HealthHandler)
    v1.GET("/users", UserListHandler)
    v1.POST("/users", CreateUserHandler)
}
```

等价于：

- `GET /api/v1/health`
- `GET /api/v1/users`
- `POST /api/v1/users`

路由组的好处：

- 路径前缀统一管理。
- 可以给整组路由添加中间件。
- API 版本控制（如 `/api/v1`、`/api/v2`）。

---

## 今天的项目结构

```
cmd/server/main.go          # 入口，创建 Gin 路由，启动服务
internal/handler/health.go  # health handler
internal/handler/health_test.go  # handler 测试
```

### main.go

```go
package main

import (
    "log"

    "github.com/CanYangTang/go_learning/internal/handler"
    "github.com/gin-gonic/gin"
)

func main() {
    gin.SetMode(gin.ReleaseMode)

    router := gin.New()

    v1 := router.Group("/api/v1")
    {
        v1.GET("/health", handler.HealthHandler)
    }

    addr := ":8080"
    log.Printf("server listening on %s", addr)
    if err := router.Run(addr); err != nil {
        log.Fatal(err)
    }
}
```

### internal/handler/health.go

```go
package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

func HealthHandler(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status":  "ok",
        "service": "go-learning",
        "version": "0.1.0",
    })
}
```

---

## Gin 测试

Gin 提供了测试工具：

```go
func TestHealthHandler(t *testing.T) {
    gin.SetMode(gin.TestMode)

    router := gin.New()
    router.GET("/api/v1/health", HealthHandler)

    req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
    recorder := httptest.NewRecorder()

    router.ServeHTTP(recorder, req)

    if recorder.Code != http.StatusOK {
        t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
    }
}
```

- `gin.SetMode(gin.TestMode)`：禁用日志输出，加速测试。
- `router.ServeHTTP(recorder, req)`：模拟请求，记录响应。

---

## 项目分层

### cmd/server

存放程序入口：

- 创建路由器。
- 注册路由组。
- 启动服务。

`main` 函数应该尽量简洁，只负责组装和启动。

### internal/handler

存放 HTTP handler：

- 处理请求。
- 返回响应。
- 不包含业务逻辑（后续会放到 service 层）。

`internal` 目录表示"内部包"，Go 工具链会阻止外部项目导入。

---

## Gin vs net/http

| 特性 | net/http | Gin |
|------|----------|-----|
| 路由分组 | 需要手动实现 | `router.Group()` |
| JSON 响应 | 手动设置头和编码 | `c.JSON()` |
| 中间件 | 需要手动实现 | 内置支持 |
| 请求绑定 | 手动解析 | `c.ShouldBindJSON()` |
| 性能 | 标准 | 更快 |

Gin 是对 `net/http` 的封装，底层仍然使用标准库。

---

## 今日验收标准

完成后应该满足：

1. Gin 依赖已添加到 `go.mod`。
2. `internal/handler/health.go` 实现 `HealthHandler`。
3. `cmd/server/main.go` 使用 Gin 并创建 `/api/v1` 路由组。
4. 测试覆盖 handler 成功响应和方法不允许。
5. `go test ./internal/handler` 通过。
6. `make test` 通过。
7. `make run` + `curl http://localhost:8080/api/v1/health` 返回 JSON。
8. 能解释 Gin 路由、路由组、handler、`gin.Context`。

---

## 可选挑战题：统一 404 响应

Issue 的可选挑战是增加统一 404 JSON 响应。

可以在 `main.go` 中添加：

```go
router.NoRoute(func(c *gin.Context) {
    c.JSON(http.StatusNotFound, gin.H{
        "error": "not found",
    })
})
```

这样所有未匹配的路由都会返回 JSON 格式的 404。

---

## 今天最容易踩的坑

### 坑 1：忘记设置 Gin 模式

生产环境应该用 `gin.ReleaseMode`：

```go
gin.SetMode(gin.ReleaseMode)
```

否则默认会输出大量日志，影响性能。

---

### 坑 2：路由组路径拼接错误

```go
v1 := router.Group("/api/v1")  // 注意不要写成 "api/v1" 或 "/api/v1/"
```

路径建议以 `/` 开头，不要以 `/` 结尾。

---

### 坑 3：测试时忘记 `gin.SetMode(gin.TestMode)`

不设置会导致测试输出大量日志。

---

### 坑 4：`gin.H` 的值类型

`gin.H` 是 `map[string]any`，值可以是任意类型，但要注意 JSON 序列化兼容性。

---

### 坑 5：`router.Run` 返回的错误

`router.Run` 在正常启动后不会返回，只有在出错时才返回错误。

所以 `log.Fatal(err)` 只在启动失败时执行。