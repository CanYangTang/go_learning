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
	todos         []model.Todo
	nextID        uint
	createErr     error
	findAllErr    error
	findByIDErr   error
	updateErr     error
	deleteErr     error
	createCalls   int
	findAllCalls  int
	findByIDCalls int
	updateCalls   int
	deleteCalls   int
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

func (f *fakeTodoRepo) FindByID(id uint) (*model.Todo, error) {
	f.findByIDCalls++
	if f.findByIDErr != nil {
		return nil, f.findByIDErr
	}
	for i := range f.todos {
		if f.todos[i].ID == id {
			todo := f.todos[i]
			return &todo, nil
		}
	}
	return nil, nil
}

func (f *fakeTodoRepo) Update(todo *model.Todo) error {
	f.updateCalls++
	if f.updateErr != nil {
		return f.updateErr
	}
	for i := range f.todos {
		if f.todos[i].ID == todo.ID {
			f.todos[i] = *todo
			return nil
		}
	}
	// Mirror GORM Save's upsert behavior so a test that skips the existence
	// check would visibly create a ghost row rather than silently no-op.
	f.todos = append(f.todos, *todo)
	return nil
}

func (f *fakeTodoRepo) Delete(id uint) (int64, error) {
	f.deleteCalls++
	if f.deleteErr != nil {
		return 0, f.deleteErr
	}
	for i := range f.todos {
		if f.todos[i].ID == id {
			f.todos = append(f.todos[:i], f.todos[i+1:]...)
			return 1, nil
		}
	}
	return 0, nil
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

func seedTodo(repo *fakeTodoRepo, title string, done bool) *model.Todo {
	todo := &model.Todo{Title: title, Done: done}
	_ = repo.Create(todo)
	return todo
}

func TestGetTodoReturnsTodo(t *testing.T) {
	repo := &fakeTodoRepo{}
	seeded := seedTodo(repo, "Learn Go", false)
	svc := NewTodoService(repo)

	todo, err := svc.GetTodo(seeded.ID)
	if err != nil {
		t.Fatalf("GetTodo failed: %v", err)
	}
	if todo == nil || todo.ID != seeded.ID || todo.Title != "Learn Go" {
		t.Fatalf("unexpected todo: %+v", todo)
	}
}

func TestGetTodoNotFound(t *testing.T) {
	repo := &fakeTodoRepo{}
	svc := NewTodoService(repo)

	todo, err := svc.GetTodo(999)

	assertAppError(t, err, "NOT_FOUND", http.StatusNotFound)
	if todo != nil {
		t.Fatalf("expected nil todo, got %+v", todo)
	}
}

func TestGetTodoWrapsRepositoryError(t *testing.T) {
	repo := &fakeTodoRepo{findByIDErr: errors.New("db is down")}
	svc := NewTodoService(repo)

	todo, err := svc.GetTodo(1)

	assertAppError(t, err, "INTERNAL_ERROR", http.StatusInternalServerError)
	if todo != nil {
		t.Fatalf("expected nil todo, got %+v", todo)
	}
}

func TestUpdateTodoUpdatesFields(t *testing.T) {
	repo := &fakeTodoRepo{}
	seeded := seedTodo(repo, "old", false)
	svc := NewTodoService(repo)

	todo, err := svc.UpdateTodo(seeded.ID, "  new title  ", true)
	if err != nil {
		t.Fatalf("UpdateTodo failed: %v", err)
	}
	if todo.Title != "new title" || !todo.Done {
		t.Fatalf("unexpected todo: %+v", todo)
	}
	stored, _ := repo.FindByID(seeded.ID)
	if stored.Title != "new title" || !stored.Done {
		t.Fatalf("repository not updated: %+v", stored)
	}
}

func TestUpdateTodoRejectsBlankTitle(t *testing.T) {
	repo := &fakeTodoRepo{}
	seeded := seedTodo(repo, "keep", false)
	svc := NewTodoService(repo)

	todo, err := svc.UpdateTodo(seeded.ID, "   ", true)

	assertAppError(t, err, "VALIDATION_ERROR", http.StatusBadRequest)
	if todo != nil {
		t.Fatalf("expected nil todo, got %+v", todo)
	}
	if repo.updateCalls != 0 {
		t.Fatalf("repo.Update should not be called for a blank title, got %d", repo.updateCalls)
	}
}

func TestUpdateTodoNotFoundDoesNotUpsert(t *testing.T) {
	repo := &fakeTodoRepo{}
	svc := NewTodoService(repo)

	todo, err := svc.UpdateTodo(42, "ghost", true)

	assertAppError(t, err, "NOT_FOUND", http.StatusNotFound)
	if todo != nil {
		t.Fatalf("expected nil todo, got %+v", todo)
	}
	if repo.updateCalls != 0 {
		t.Fatalf("repo.Update must not be called for a missing id (would upsert a ghost row), got %d", repo.updateCalls)
	}
	if len(repo.todos) != 0 {
		t.Fatalf("no row should have been created, got %+v", repo.todos)
	}
}

func TestUpdateTodoWrapsFindError(t *testing.T) {
	repo := &fakeTodoRepo{findByIDErr: errors.New("db is down")}
	svc := NewTodoService(repo)

	_, err := svc.UpdateTodo(1, "title", false)

	assertAppError(t, err, "INTERNAL_ERROR", http.StatusInternalServerError)
}

func TestUpdateTodoWrapsUpdateError(t *testing.T) {
	repo := &fakeTodoRepo{updateErr: errors.New("db is down")}
	seeded := seedTodo(repo, "old", false)
	svc := NewTodoService(repo)

	_, err := svc.UpdateTodo(seeded.ID, "new", true)

	assertAppError(t, err, "INTERNAL_ERROR", http.StatusInternalServerError)
}

func TestDeleteTodoDeletes(t *testing.T) {
	repo := &fakeTodoRepo{}
	seeded := seedTodo(repo, "bye", false)
	svc := NewTodoService(repo)

	if err := svc.DeleteTodo(seeded.ID); err != nil {
		t.Fatalf("DeleteTodo failed: %v", err)
	}
	if repo.deleteCalls != 1 {
		t.Fatalf("repo.Delete called %d times, want 1", repo.deleteCalls)
	}
	if stored, _ := repo.FindByID(seeded.ID); stored != nil {
		t.Fatalf("todo should be gone, got %+v", stored)
	}
}

func TestDeleteTodoNotFound(t *testing.T) {
	repo := &fakeTodoRepo{}
	svc := NewTodoService(repo)

	err := svc.DeleteTodo(999)

	assertAppError(t, err, "NOT_FOUND", http.StatusNotFound)
	// Choice B: Delete IS called; RowsAffected == 0 is what signals not-found.
	if repo.deleteCalls != 1 {
		t.Fatalf("repo.Delete called %d times, want 1", repo.deleteCalls)
	}
	if repo.findByIDCalls != 0 {
		t.Fatalf("DeleteTodo should not call FindByID under choice B, got %d", repo.findByIDCalls)
	}
}

func TestDeleteTodoWrapsDeleteError(t *testing.T) {
	repo := &fakeTodoRepo{deleteErr: errors.New("db is down")}
	seeded := seedTodo(repo, "bye", false)
	svc := NewTodoService(repo)

	err := svc.DeleteTodo(seeded.ID)

	assertAppError(t, err, "INTERNAL_ERROR", http.StatusInternalServerError)
}
