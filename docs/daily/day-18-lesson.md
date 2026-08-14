# Day 18 教案：GORM 模型和 repository

## 学习目标

学完今天，需要能够做到：

1. 理解 GORM 的基本概念和使用方式。
2. 使用 GORM 定义模型和表映射。
3. 使用 GORM 进行创建、查询、更新、删除。
4. 理解 `gorm.Model` 和自定义字段的区别。
5. 编写 repository 层测试。
6. 了解事务的基本使用。

---

## Day 18 的位置

Day 17 学了 `database/sql` 和 MySQL 基础 CRUD。

Day 18 会引入 GORM，把手写 SQL 的 repository 改成 ORM 风格。

Day 19 会加入日志、认证、CORS 中间件。

---

## GORM 是什么

GORM 是 Go 里最流行的 ORM 框架之一。

ORM（Object Relational Mapping）把数据库表和 Go struct 映射起来，让开发者用 struct 和方法操作数据库，而不是手写大量 SQL。

特点：

- 约定优于配置。
- 支持 CRUD、关联、事务、预加载。
- 支持 MySQL、PostgreSQL、SQLite 等多种数据库。

导入：

```go
import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)
```

---

## GORM 连接数据库

```go
dsn := "root:password@tcp(127.0.0.1:3306)/go_learning?charset=utf8mb4&parseTime=True&loc=Local"
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
if err != nil {
    log.Fatal(err)
}
```

- `gorm.Open` 创建 GORM DB 对象。
- `mysql.Open(dsn)` 提供 MySQL 驱动。
- `gorm.Config{}` 可以传入 GORM 配置。

---

## GORM 模型

GORM 通过 struct 定义模型。

```go
type Todo struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Title     string         `gorm:"size:255;not null" json:"title"`
    Done      bool           `gorm:"default:false" json:"done"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

### `gorm.Model`

GORM 提供内置结构：

```go
type Model struct {
    ID        uint           `gorm:"primaryKey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

如果直接嵌入 `gorm.Model`：

```go
type Todo struct {
    gorm.Model
    Title string `gorm:"size:255;not null"`
    Done  bool   `gorm:"default:false"`
}
```

优点：

- 自动包含常见字段。
- 支持软删除。

缺点：

- 字段固定，不够灵活。
- JSON tag 不好完全控制。

今天建议显式定义字段，便于学习和测试。

---

## GORM 标签

常见标签：

- `primaryKey`：主键。
- `size:255`：字段长度。
- `not null`：非空。
- `default:false`：默认值。
- `index`：建索引。

示例：

```go
Title string `gorm:"size:255;not null" json:"title"`
```

---

## AutoMigrate

GORM 可以自动建表：

```go
if err := db.AutoMigrate(&model.Todo{}); err != nil {
    return err
}
```

`AutoMigrate` 会：

- 创建表（如果不存在）。
- 自动增加缺失字段。
- 不会删除字段或表。

适合开发阶段，不建议直接依赖它做完整生产迁移。

---

## repository 层

repository 负责数据访问。

今天建议定义：

```go
type TodoRepository struct {
    db *gorm.DB
}
```

方法：

- `Create(todo *model.Todo) error`
- `FindByID(id uint) (*model.Todo, error)`
- `FindAll() ([]model.Todo, error)`
- `Update(todo *model.Todo) error`
- `Delete(id uint) error`

---

## GORM CRUD 示例

### Create

```go
err := r.db.Create(todo).Error
```

`Create` 会把 `todo.ID` 自动回填。

### FindByID

```go
var todo model.Todo
err := r.db.First(&todo, id).Error
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, nil
}
```

### FindAll

```go
var todos []model.Todo
err := r.db.Find(&todos).Error
```

### Update

```go
err := r.db.Save(todo).Error
```

或者：

```go
err := r.db.Model(&model.Todo{}).Where("id = ?", todo.ID).Updates(todo).Error
```

### Delete

```go
err := r.db.Delete(&model.Todo{}, id).Error
```

---

## 事务

事务用于把多个数据库操作作为一个原子整体执行。

### 基本写法

```go
tx := db.Begin()
if tx.Error != nil {
    return tx.Error
}

defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
        panic(r)
    }
}()

if err := tx.Create(&todo).Error; err != nil {
    tx.Rollback()
    return err
}

if err := tx.Commit().Error; err != nil {
    return err
}
```

更常见的写法是使用 `db.Transaction(func(tx *gorm.DB) error { ... })`。

---

## 今日代码结构

```
internal/model/todo.go          # GORM Todo 模型
internal/repository/todo.go     # GORM Todo repository
internal/repository/todo_test.go
```

---

## 今日验收标准

完成后应该满足：

1. `internal/model/todo.go` 使用 GORM 定义 Todo 模型。
2. `internal/repository/todo.go` 使用 GORM 实现 CRUD。
3. 至少写 3 组 repository 测试。
4. `go test ./internal/repository` 通过。
5. `make test` 通过。
6. 能解释 GORM 模型、`AutoMigrate`、`ErrRecordNotFound`、事务。

---

## 可选挑战题：事务示例

可以增加一个事务方法：

```go
func (r *TodoRepository) CreateAndMarkDone(todo *model.Todo) error
```

流程：

1. 创建 todo。
2. 更新 done 状态。
3. 任一步失败则回滚。

使用：

```go
db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&todo).Error; err != nil {
        return err
    }
    todo.Done = true
    if err := tx.Save(&todo).Error; err != nil {
        return err
    }
    return nil
})
```

---

## 今天最容易踩的坑

### 坑 1：忘记导入 MySQL driver

GORM 的 MySQL 驱动也需要导入：

```go
import "gorm.io/driver/mysql"
```

---

### 坑 2：`ErrRecordNotFound` 处理

GORM 找不到记录时返回 `gorm.ErrRecordNotFound`，不是 `sql.ErrNoRows`。

---

### 坑 3：`AutoMigrate` 不等于完整迁移

它不会删除字段，也不会处理复杂迁移逻辑。

---

### 坑 4：`Save` 可能更新所有字段

`Save` 会保存整个对象，使用时要注意零值覆盖问题。

---

### 坑 5：事务中忘记返回错误

在 `db.Transaction` 回调里，必须把错误 return 出去，否则事务会提交。
