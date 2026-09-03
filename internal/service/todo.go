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
