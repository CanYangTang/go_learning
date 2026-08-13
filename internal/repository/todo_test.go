package repository

import (
	"database/sql"
	"os"
	"testing"

	"github.com/CanYangTang/go_learning/internal/config"
	"github.com/CanYangTang/go_learning/internal/model"
)

var testDB *sql.DB

// TestMain sets up the database connection for all tests.
// Run: docker-compose -f deployments/docker-compose.yml up -d
// Then: go test ./internal/repository -tags=integration
func TestMain(m *testing.M) {
	// Skip integration tests if not running with -tags=integration
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		os.Exit(m.Run())
	}

	cfg := &config.DatabaseConfig{
		Host:         "127.0.0.1",
		Port:         3306,
		Username:     "root",
		Password:     "password",
		Database:     "go_learning",
		MaxOpenConns: 10,
		MaxIdleConns: 10,
	}

	var err error
	testDB, err = config.Connect(cfg)
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	// Create table for testing
	_, err = testDB.Exec(`
		CREATE TABLE IF NOT EXISTS todos (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			done BOOLEAN DEFAULT FALSE
		)
	`)
	if err != nil {
		panic("failed to create table: " + err.Error())
	}

	code := m.Run()

	// Cleanup
	testDB.Exec("DROP TABLE IF EXISTS todos")
	testDB.Close()
	os.Exit(code)
}

func TestTodoRepository_Create(t *testing.T) {
	if testDB == nil {
		t.Skip("integration tests require database connection")
	}

	repo := NewTodoRepository(testDB)

	todo := &model.Todo{
		Title: "Test Todo",
		Done:  false,
	}

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

	repo := NewTodoRepository(testDB)

	// Create a todo first
	todo := &model.Todo{Title: "Find Test", Done: false}
	_ = repo.Create(todo)

	// Find it
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

func TestTodoRepository_FindAll(t *testing.T) {
	if testDB == nil {
		t.Skip("integration tests require database connection")
	}

	repo := NewTodoRepository(testDB)

	// Create some todos
	_ = repo.Create(&model.Todo{Title: "Todo 1", Done: false})
	_ = repo.Create(&model.Todo{Title: "Todo 2", Done: true})

	todos, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	if len(todos) < 2 {
		t.Fatalf("expected at least 2 todos, got %d", len(todos))
	}
}

func TestTodoRepository_Update(t *testing.T) {
	if testDB == nil {
		t.Skip("integration tests require database connection")
	}

	repo := NewTodoRepository(testDB)

	todo := &model.Todo{Title: "Update Test", Done: false}
	_ = repo.Create(todo)

	todo.Done = true
	err := repo.Update(todo)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify
	found, _ := repo.FindByID(todo.ID)
	if !found.Done {
		t.Fatalf("expected Done to be true")
	}
}

func TestTodoRepository_Delete(t *testing.T) {
	if testDB == nil {
		t.Skip("integration tests require database connection")
	}

	repo := NewTodoRepository(testDB)

	todo := &model.Todo{Title: "Delete Test", Done: false}
	_ = repo.Create(todo)

	err := repo.Delete(todo.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify
	found, _ := repo.FindByID(todo.ID)
	if found != nil {
		t.Fatalf("expected todo to be deleted")
	}
}