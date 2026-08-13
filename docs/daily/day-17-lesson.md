# Day 17 教案：MySQL 连接和基础 CRUD

## 学习目标

学完今天，需要能够做到：

1. 理解 Go 的 `database/sql` 包基础用法。
2. 使用 `go-sql-driver/mysql` 驱动连接 MySQL。
3. 配置数据库连接（DSN、连接池）。
4. 使用 Docker Compose 启动 MySQL 开发环境。
5. 实现基础 CRUD 操作（Create、Read、Update、Delete）。
6. 编写数据库操作的测试。

---

## Day 17 的位置

Day 16 实现了 TODO 的创建和列表接口，数据暂时存在内存中。

Day 17 会把数据持久化到 MySQL。

Day 18 会引入 GORM 简化数据库操作。

---

## database/sql 包

Go 标准库 `database/sql` 提供了数据库操作的通用接口。

特点：

- 面向接口设计，支持多种数据库驱动。
- 提供连接池、预处理、事务等能力。
- 不依赖具体数据库，驱动由第三方提供。

导入：

```go
import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql" // MySQL 驱动
)
```

注意：驱动用 `_` 导入，只执行 `init()` 注册驱动，不直接使用。

---

## DSN（数据源名称）

DSN 是连接数据库的配置字符串。

MySQL 的 DSN 格式：

```text
username:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
```

示例：

```text
root:password@tcp(127.0.0.1:3306)/go_learning?charset=utf8mb4&parseTime=True&loc=Local
```

参数说明：

- `charset=utf8mb4`：字符集。
- `parseTime=True`：自动把 `DATETIME` 转成 `time.Time`。
- `loc=Local`：时区设置。

---

## 连接 MySQL

### sql.Open

```go
db, err := sql.Open("mysql", dsn)
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

- `"mysql"`：驱动名称。
- `dsn`：数据源名称。
- `sql.Open` 不实际连接，只是验证参数并创建 `*sql.DB` 对象。

### db.Ping

验证连接：

```go
if err := db.Ping(); err != nil {
    log.Fatal(err)
}
```

`Ping` 才会真正尝试连接数据库。

---

## 连接池配置

`*sql.DB` 内置连接池，可以配置：

```go
db.SetMaxOpenConns(25)     // 最大打开连接数
db.SetMaxIdleConns(25)     // 最大空闲连接数
db.SetConnMaxLifetime(5 * time.Minute) // 连接最大存活时间
```

建议：

- `MaxOpenConns`：根据数据库承载能力设置，通常 10-100。
- `MaxIdleConns`：与 `MaxOpenConns` 相近，避免频繁创建/销毁连接。
- `ConnMaxLifetime`：小于数据库的 `wait_timeout`。

---

## 配置结构体

定义数据库配置：

```go
type Config struct {
    Host         string
    Port         int
    Username     string
    Password     string
    Database     string
    MaxOpenConns int
    MaxIdleConns int
}

func (c *Config) DSN() string {
    return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        c.Username, c.Password, c.Host, c.Port, c.Database)
}
```

---

## Docker Compose 启动 MySQL

创建 `docker-compose.yml`：

```yaml
version: '3.8'
services:
  mysql:
    image: mysql:8.0
    container_name: go_learning_mysql
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: go_learning
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql

volumes:
  mysql_data:
```

启动：

```bash
docker-compose up -d
```

停止：

```bash
docker-compose down
```

---

## CRUD 操作

### Create（插入）

```go
result, err := db.Exec(
    "INSERT INTO todos (title, done) VALUES (?, ?)",
    "Learn Go", false,
)
if err != nil {
    return err
}
id, _ := result.LastInsertId()
```

- `db.Exec` 执行插入/更新/删除。
- `?` 是占位符，防止 SQL 注入。
- `LastInsertId()` 返回自增 ID。

### Read（查询单行）

```go
var todo Todo
err := db.QueryRow(
    "SELECT id, title, done FROM todos WHERE id = ?",
    1,
).Scan(&todo.ID, &todo.Title, &todo.Done)
if err == sql.ErrNoRows {
    // 没找到
}
```

- `QueryRow` 查询单行。
- `Scan` 把列值映射到变量。
- `sql.ErrNoRows` 表示没有查到数据。

### Read（查询多行）

```go
rows, err := db.Query("SELECT id, title, done FROM todos")
if err != nil {
    return err
}
defer rows.Close()

var todos []Todo
for rows.Next() {
    var todo Todo
    if err := rows.Scan(&todo.ID, &todo.Title, &todo.Done); err != nil {
        return err
    }
    todos = append(todos, todo)
}
if err := rows.Err(); err != nil {
    return err
}
```

- `db.Query` 查询多行，返回 `*sql.Rows`。
- `rows.Next()` 迭代每一行。
- 必须调用 `rows.Close()`，通常用 `defer`。
- 最后检查 `rows.Err()` 捕获迭代过程中的错误。

### Update（更新）

```go
result, err := db.Exec(
    "UPDATE todos SET done = ? WHERE id = ?",
    true, 1,
)
if err != nil {
    return err
}
affected, _ := result.RowsAffected()
```

- `RowsAffected()` 返回影响行数。

### Delete（删除）

```go
result, err := db.Exec("DELETE FROM todos WHERE id = ?", 1)
if err != nil {
    return err
}
affected, _ := result.RowsAffected()
```

---

## 项目结构

```
internal/config/database.go     # 数据库配置和连接
internal/model/todo.go          # Todo 模型
internal/repository/todo.go     # Todo 数据访问
deployments/docker-compose.yml  # MySQL 开发环境
```

---

## 数据库配置

`internal/config/database.go`：

```go
package config

import (
    "database/sql"
    "fmt"

    _ "github.com/go-sql-driver/mysql"
)

type DatabaseConfig struct {
    Host         string
    Port         int
    Username     string
    Password     string
    Database     string
    MaxOpenConns int
    MaxIdleConns int
}

func (c *DatabaseConfig) DSN() string {
    return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        c.Username, c.Password, c.Host, c.Port, c.Database)
}

func Connect(cfg *DatabaseConfig) (*sql.DB, error) {
    db, err := sql.Open("mysql", cfg.DSN())
    if err != nil {
        return nil, err
    }

    db.SetMaxOpenConns(cfg.MaxOpenConns)
    db.SetMaxIdleConns(cfg.MaxIdleConns)

    if err := db.Ping(); err != nil {
        return nil, err
    }

    return db, nil
}
```

---

## Repository 层

`internal/repository/todo.go`：

```go
package repository

import (
    "database/sql"
    "github.com/CanYangTang/go_learning/internal/model"
)

type TodoRepository struct {
    db *sql.DB
}

func NewTodoRepository(db *sql.DB) *TodoRepository {
    return &TodoRepository{db: db}
}

func (r *TodoRepository) Create(todo *model.Todo) error {
    result, err := r.db.Exec(
        "INSERT INTO todos (title, done) VALUES (?, ?)",
        todo.Title, todo.Done,
    )
    if err != nil {
        return err
    }
    id, _ := result.LastInsertId()
    todo.ID = uint(id)
    return nil
}

func (r *TodoRepository) FindByID(id uint) (*model.Todo, error) {
    todo := &model.Todo{}
    err := r.db.QueryRow(
        "SELECT id, title, done FROM todos WHERE id = ?",
        id,
    ).Scan(&todo.ID, &todo.Title, &todo.Done)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return todo, err
}

func (r *TodoRepository) FindAll() ([]model.Todo, error) {
    rows, err := r.db.Query("SELECT id, title, done FROM todos")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var todos []model.Todo
    for rows.Next() {
        var todo model.Todo
        if err := rows.Scan(&todo.ID, &todo.Title, &todo.Done); err != nil {
            return nil, err
        }
        todos = append(todos, todo)
    }
    return todos, rows.Err()
}

func (r *TodoRepository) Update(todo *model.Todo) error {
    _, err := r.db.Exec(
        "UPDATE todos SET title = ?, done = ? WHERE id = ?",
        todo.Title, todo.Done, todo.ID,
    )
    return err
}

func (r *TodoRepository) Delete(id uint) error {
    _, err := r.db.Exec("DELETE FROM todos WHERE id = ?", id)
    return err
}
```

---

## 今日验收标准

完成后应该满足：

1. Docker Compose 能启动 MySQL。
2. `internal/config/database.go` 能连接 MySQL。
3. `internal/repository/todo.go` 实现 CRUD。
4. 至少写 3 组测试（Create、FindByID、FindAll）。
5. `go test ./internal/repository` 通过（需要数据库运行）。
6. 能解释 DSN、`sql.Open`、`db.Ping`、连接池配置。

---

## 可选挑战题：集成测试

由于数据库测试需要真实连接，可以使用 Docker Compose 启动的 MySQL 运行集成测试。

测试前确保：

```bash
docker-compose up -d
```

测试后可以清理：

```bash
docker-compose down -v
```

---

## 今天最容易踩的坑

### 坑 1：忘记导入驱动

必须导入 MySQL 驱动：

```go
import _ "github.com/go-sql-driver/mysql"
```

否则 `sql.Open("mysql", ...)` 会报 "unknown driver"。

---

### 坑 2：DSN 格式错误

常见错误：

- 密码包含特殊字符（如 `@`、`#`），需要 URL 编码。
- 忘记加 `parseTime=True`，查询 `DATETIME` 会失败。

---

### 坑 3：忘记 `rows.Close()`

`db.Query` 返回的 `*sql.Rows` 必须关闭，否则连接泄漏。

```go
rows, err := db.Query(...)
if err != nil { ... }
defer rows.Close()  // 必须
```

---

### 坑 4：`sql.ErrNoRows` 处理

`QueryRow` 找不到数据时返回 `sql.ErrNoRows`，需要特殊处理：

```go
if err == sql.ErrNoRows {
    return nil, nil  // 或返回特定业务错误
}
```

---

### 坑 5：事务未提交或回滚

```go
tx, err := db.Begin()
if err != nil { ... }
defer tx.Rollback()  // 安全做法，确保出错时回滚

// ... 执行操作

if err := tx.Commit(); err != nil { ... }
```

`defer tx.Rollback()` 在 `Commit()` 成功后会变成无操作。