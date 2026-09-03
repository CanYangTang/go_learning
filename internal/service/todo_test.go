package service

import (
	"errors"
	"net/http"
	"testing"

	"github.com/CanYangTang/go_learning/internal/model"
	"github.com/CanYangTang/go_learning/internal/repository"
	"github.com/CanYangTang/go_learning/pkg/apperror"
)

// The real repository must satisfy the interface declared by the service.
var _ TodoRepository = (*repository.TodoRepository)(nil)

// fakeTodoRepo is an in-memory stand-in for the database.
type fakeTodoRepo struct {
	todos        []model.Todo
	nextID       uint
	createErr    error
	findAllErr   error
	createCalls  int
	findAllCalls int
}

var _ TodoRepository = (*fakeTodoRepo)(nil)

func (f *fakeTodoRepo) Create(todo *model.Todo) error {
	f.createCalls++
	if f.createErr != nil {
		return f.createErr
	}
	f.nextID++
	todo.ID = f.nextID
	f.todos = append(f.todos, *todo)
	return nil
}

func (f *fakeTodoRepo) FindAll() ([]model.Todo, error) {
	f.findAllCalls++
	if f.findAllErr != nil {
		return nil, f.findAllErr
	}
	return f.todos, nil
}

func assertAppError(t *testing.T, err error, wantCode string, wantStatus int) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected an error, got nil")
	}

	var appErr apperror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("error %v is not an apperror.Error", err)
	}
	if appErr.Code != wantCode {
		t.Fatalf("code = %q, want %q", appErr.Code, wantCode)
	}
	if appErr.StatusCode != wantStatus {
		t.Fatalf("status = %d, want %d", appErr.StatusCode, wantStatus)
	}
}

func TestCreateTodoStoresTodoAndFillsID(t *testing.T) {
	repo := &fakeTodoRepo{}
	svc := NewTodoService(repo)

	todo, err := svc.CreateTodo("Learn Go")
	if err != nil {
		t.Fatalf("CreateTodo failed: %v", err)
	}
	if todo == nil {
		t.Fatalf("expected a todo, got nil")
	}
	if todo.ID == 0 {
		t.Fatalf("todo.ID should be filled in by the repository")
	}
	if todo.Title != "Learn Go" {
		t.Fatalf("title = %q, want %q", todo.Title, "Learn Go")
	}
	if todo.Done {
		t.Fatalf("a new todo should not be done")
	}
	if repo.createCalls != 1 {
		t.Fatalf("repo.Create called %d times, want 1", repo.createCalls)
	}
}

func TestCreateTodoTrimsTitle(t *testing.T) {
	repo := &fakeTodoRepo{}
	svc := NewTodoService(repo)

	todo, err := svc.CreateTodo("  Learn Go  ")
	if err != nil {
		t.Fatalf("CreateTodo failed: %v", err)
	}
	if todo == nil {
		t.Fatalf("expected a todo, got nil")
	}
	if todo.Title != "Learn Go" {
		t.Fatalf("title = %q, want %q", todo.Title, "Learn Go")
	}
}

func TestCreateTodoRejectsBlankTitle(t *testing.T) {
	cases := map[string]string{
		"empty":      "",
		"spaces":     "   ",
		"whitespace": "\t\n ",
	}

	for name, title := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &fakeTodoRepo{}
			svc := NewTodoService(repo)

			todo, err := svc.CreateTodo(title)

			assertAppError(t, err, "VALIDATION_ERROR", http.StatusBadRequest)
			if todo != nil {
				t.Fatalf("expected nil todo, got %+v", todo)
			}
			if repo.createCalls != 0 {
				t.Fatalf("repo.Create should not be called for an invalid title, got %d calls", repo.createCalls)
			}
		})
	}
}

func TestCreateTodoWrapsRepositoryError(t *testing.T) {
	repo := &fakeTodoRepo{createErr: errors.New("db is down")}
	svc := NewTodoService(repo)

	todo, err := svc.CreateTodo("Learn Go")

	assertAppError(t, err, "INTERNAL_ERROR", http.StatusInternalServerError)
	if todo != nil {
		t.Fatalf("expected nil todo, got %+v", todo)
	}
}

func TestListTodosReturnsEmptySliceWhenNoRows(t *testing.T) {
	repo := &fakeTodoRepo{}
	svc := NewTodoService(repo)

	todos, err := svc.ListTodos()
	if err != nil {
		t.Fatalf("ListTodos failed: %v", err)
	}
	if todos == nil {
		t.Fatalf("expected an empty slice, got nil")
	}
	if len(todos) != 0 {
		t.Fatalf("len(todos) = %d, want 0", len(todos))
	}
	if repo.findAllCalls != 1 {
		t.Fatalf("repo.FindAll called %d times, want 1", repo.findAllCalls)
	}
}

func TestListTodosReturnsStoredTodos(t *testing.T) {
	repo := &fakeTodoRepo{}
	svc := NewTodoService(repo)

	if _, err := svc.CreateTodo("Todo 1"); err != nil {
		t.Fatalf("CreateTodo failed: %v", err)
	}
	if _, err := svc.CreateTodo("Todo 2"); err != nil {
		t.Fatalf("CreateTodo failed: %v", err)
	}

	todos, err := svc.ListTodos()
	if err != nil {
		t.Fatalf("ListTodos failed: %v", err)
	}
	if len(todos) != 2 {
		t.Fatalf("len(todos) = %d, want 2", len(todos))
	}
	if todos[0].Title != "Todo 1" || todos[1].Title != "Todo 2" {
		t.Fatalf("unexpected todos: %+v", todos)
	}
}

func TestListTodosWrapsRepositoryError(t *testing.T) {
	repo := &fakeTodoRepo{findAllErr: errors.New("db is down")}
	svc := NewTodoService(repo)

	todos, err := svc.ListTodos()

	assertAppError(t, err, "INTERNAL_ERROR", http.StatusInternalServerError)
	if todos != nil {
		t.Fatalf("expected nil todos, got %+v", todos)
	}
}
