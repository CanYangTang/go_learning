package repository

import (
	"os"
	"testing"

	"github.com/CanYangTang/go_learning/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var testDB *gorm.DB

// TestMain sets up the database connection for all tests.
// Run: docker compose -f deployments/docker-compose.yml up -d
// Then: RUN_INTEGRATION_TESTS=true go test ./internal/repository -v
func TestMain(m *testing.M) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		os.Exit(m.Run())
	}

	dsn := "root:password@tcp(127.0.0.1:3306)/go_learning?charset=utf8mb4&parseTime=True&loc=Local"

	var err error
	testDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	if err := testDB.AutoMigrate(&model.Todo{}); err != nil {
		panic("failed to migrate database: " + err.Error())
	}

	code := m.Run()

	sqlDB, err := testDB.DB()
	if err == nil {
		testDB.Exec("DROP TABLE IF EXISTS todos")
		sqlDB.Close()
	}
	os.Exit(code)
}

func cleanupTodos(t *testing.T) {
	t.Helper()
	if testDB == nil {
		return
	}
	if err := testDB.Exec("DELETE FROM todos").Error; err != nil {
		t.Fatalf("cleanup todos failed: %v", err)
	}
}

func TestTodoRepository_Create(t *testing.T) {
	if testDB == nil {
		t.Skip("integration tests require database connection")
	}
	cleanupTodos(t)

	repo := NewTodoRepository(testDB)

	todo := &model.Todo{Title: "Test Todo", Done: false}

	err := repo.Create(todo)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if todo.ID == 0 {
		t.Fatalf("todo.ID should be set after Create")
	}
}

func TestTodoRepository_FindByID(t *testing.T) {
	if testDB == nil {
		t.Skip("integration tests require database connection")
	}
	cleanupTodos(t)

	repo := NewTodoRepository(testDB)

	todo := &model.Todo{Title: "Find Test", Done: false}
	if err := repo.Create(todo); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByID(todo.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if found == nil {
		t.Fatalf("expected to find todo with ID %d", todo.ID)
	}

	if found.Title != todo.Title {
		t.Fatalf("title = %q, want %q", found.Title, todo.Title)
	}
}

func TestTodoRepository_FindByIDNotFound(t *testing.T) {
	if testDB == nil {
		t.Skip("integration tests require database connection")
	}
	cleanupTodos(t)

	repo := NewTodoRepository(testDB)

	found, err := repo.FindByID(999999)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if found != nil {
		t.Fatalf("expected nil todo, got %+v", found)
	}
}

func TestTodoRepository_FindAll(t *testing.T) {
	if testDB == nil {
		t.Skip("integration tests require database connection")
	}
	cleanupTodos(t)

	repo := NewTodoRepository(testDB)

	if err := repo.Create(&model.Todo{Title: "Todo 1", Done: false}); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if err := repo.Create(&model.Todo{Title: "Todo 2", Done: true}); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	todos, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	if len(todos) != 2 {
		t.Fatalf("expected 2 todos, got %d", len(todos))
	}
}

func TestTodoRepository_Update(t *testing.T) {
	if testDB == nil {
		t.Skip("integration tests require database connection")
	}
	cleanupTodos(t)

	repo := NewTodoRepository(testDB)

	todo := &model.Todo{Title: "Update Test", Done: false}
	if err := repo.Create(todo); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	todo.Title = "Updated Title"
	todo.Done = true
	if err := repo.Update(todo); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, err := repo.FindByID(todo.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Title != "Updated Title" || !found.Done {
		t.Fatalf("todo not updated correctly: %+v", found)
	}
}

func TestTodoRepository_Delete(t *testing.T) {
	if testDB == nil {
		t.Skip("integration tests require database connection")
	}
	cleanupTodos(t)

	repo := NewTodoRepository(testDB)

	todo := &model.Todo{Title: "Delete Test", Done: false}
	if err := repo.Create(todo); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := repo.Delete(todo.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	found, err := repo.FindByID(todo.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found != nil {
		t.Fatalf("expected todo to be deleted")
	}
}

func TestTodoRepository_CreateAndMarkDone(t *testing.T) {
	if testDB == nil {
		t.Skip("integration tests require database connection")
	}
	cleanupTodos(t)

	repo := NewTodoRepository(testDB)

	todo := &model.Todo{Title: "Transaction Test", Done: false}
	if err := repo.CreateAndMarkDone(todo); err != nil {
		t.Fatalf("CreateAndMarkDone failed: %v", err)
	}

	if todo.ID == 0 {
		t.Fatalf("todo.ID should be set after CreateAndMarkDone")
	}

	found, err := repo.FindByID(todo.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found == nil || !found.Done {
		t.Fatalf("expected todo to be marked done, got %+v", found)
	}
}