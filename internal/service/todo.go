package service

import (
	"strings"

	"github.com/CanYangTang/go_learning/internal/model"
	"github.com/CanYangTang/go_learning/pkg/apperror"
)

// TodoRepository is the persistence contract the service depends on.
// Note: this interface lives in the consumer package (service), not in repository.
// It only declares the methods the service actually uses.
type TodoRepository interface {
	Create(todo *model.Todo) error
	FindAll() ([]model.Todo, error)
	FindByID(id uint) (*model.Todo, error)
	Update(todo *model.Todo) error
	Delete(id uint) (int64, error)
}

// TodoService holds the business rules for todos.
type TodoService struct {
	repo TodoRepository
}

// NewTodoService creates a new TodoService.
func NewTodoService(repo TodoRepository) *TodoService {
	return &TodoService{
		repo: repo,
	}
}

// CreateTodo validates the title and stores a new todo.
func (s *TodoService) CreateTodo(title string) (*model.Todo, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, apperror.Validation("title is required")
	}

	todo := &model.Todo{Title: title}
	if err := s.repo.Create(todo); err != nil {
		return nil, apperror.Internal("create todo failed")
	}
	return todo, nil
}

// ListTodos returns all todos.
func (s *TodoService) ListTodos() ([]model.Todo, error) {
	todos, err := s.repo.FindAll()
	if err != nil {
		return nil, apperror.Internal("list todos failed")
	}
	if todos == nil {
		// Normalize to an empty slice so the JSON response is [] instead of null.
		todos = []model.Todo{}
	}
	return todos, nil
}

// GetTodo returns a single todo by ID.
//
// TODO: implement.
//   - Call s.repo.FindByID(id).
//   - On error, return nil, apperror.Internal("get todo failed").
//   - repo.FindByID returns (nil, nil) when the row does not exist: in that
//     case return nil, apperror.NotFound("todo not found").
//   - Otherwise return the todo.
func (s *TodoService) GetTodo(id uint) (*model.Todo, error) {
	todo, err := s.repo.FindByID(id)
	if err != nil {
		return nil, apperror.Internal("get todo failed")
	}
	if todo == nil {
		return nil, apperror.NotFound("todo not found")
	}
	return todo, nil
}

// UpdateTodo does a full update of a todo's title and done flag.
//
// TODO: implement.
//   - TrimSpace(title); if empty return nil, apperror.Validation("title is required").
//   - Call s.repo.FindByID(id) to confirm the row exists FIRST:
//   - error  -> nil, apperror.Internal("update todo failed")
//   - nil    -> nil, apperror.NotFound("todo not found")
//     This existence check is what stops repo.Update's Save() from upserting a
//     brand-new row when the id does not exist (see the lesson's Save-upsert坑).
//   - Set todo.Title = title and todo.Done = done, then s.repo.Update(todo);
//     on error return nil, apperror.Internal("update todo failed").
//   - Return the updated todo.
func (s *TodoService) UpdateTodo(id uint, title string, done bool) (*model.Todo, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, apperror.Validation("title is required")
	}

	todo, err := s.repo.FindByID(id)
	if err != nil {
		return nil, apperror.Internal("update todo failed")
	}
	if todo == nil {
		return nil, apperror.NotFound("todo not found")
	}

	todo.Title = title
	todo.Done = done
	if err := s.repo.Update(todo); err != nil {
		return nil, apperror.Internal("update todo failed")
	}
	return todo, nil
}

// DeleteTodo deletes a todo by ID.
//
// Choice B (see day-24 lesson): instead of a prior FindByID existence check,
// we let repo.Delete report RowsAffected. GORM's Delete on a missing id is a
// no-op (no error, RowsAffected == 0), so RowsAffected == 0 IS the not-found
// signal. This keeps delete to a single query rather than find-then-delete.
func (s *TodoService) DeleteTodo(id uint) error {
	rows, err := s.repo.Delete(id)
	if err != nil {
		return apperror.Internal("delete todo failed")
	}
	if rows == 0 {
		return apperror.NotFound("todo not found")
	}
	return nil
}
