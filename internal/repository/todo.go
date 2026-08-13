package repository

import (
	"database/sql"

	"github.com/CanYangTang/go_learning/internal/model"
)

// TodoRepository handles database operations for todos.
type TodoRepository struct {
	db *sql.DB
}

// NewTodoRepository creates a new TodoRepository.
func NewTodoRepository(db *sql.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

// Create inserts a new todo and sets the ID.
func (r *TodoRepository) Create(todo *model.Todo) error {
	result, err := r.db.Exec(
		"INSERT INTO todos (title, done) VALUES (?, ?)",
		todo.Title, todo.Done,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	todo.ID = uint(id)
	return nil
}

// FindByID retrieves a todo by ID. Returns nil if not found.
func (r *TodoRepository) FindByID(id uint) (*model.Todo, error) {
	todo := &model.Todo{}
	err := r.db.QueryRow(
		"SELECT id, title, done FROM todos WHERE id = ?",
		id,
	).Scan(&todo.ID, &todo.Title, &todo.Done)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return todo, nil
}

// FindAll retrieves all todos.
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return todos, nil
}

// Update modifies an existing todo.
func (r *TodoRepository) Update(todo *model.Todo) error {
	_, err := r.db.Exec(
		"UPDATE todos SET title = ?, done = ? WHERE id = ?",
		todo.Title, todo.Done, todo.ID,
	)
	return err
}

// Delete removes a todo by ID.
func (r *TodoRepository) Delete(id uint) error {
	_, err := r.db.Exec("DELETE FROM todos WHERE id = ?", id)
	return err
}
