package repository

import (
	"github.com/CanYangTang/go_learning/internal/model"
	"gorm.io/gorm"
)

// TodoRepository handles database operations for todos using GORM.
type TodoRepository struct {
	db *gorm.DB
}

// NewTodoRepository creates a new TodoRepository.
func NewTodoRepository(db *gorm.DB) *TodoRepository {
	// TODO: implement
	return &TodoRepository{
		db: db,
	}
}

// Create inserts a new todo and sets the ID.
func (r *TodoRepository) Create(todo *model.Todo) error {
	// TODO: implement using r.db.Create(todo).Error
	err := r.db.Create(todo).Error
	if err != nil {
		return err
	}
	return nil
}

// FindByID retrieves a todo by ID. Returns nil if not found.
func (r *TodoRepository) FindByID(id uint) (*model.Todo, error) {
	// TODO: implement using r.db.First(&todo, id).Error
	// Handle gorm.ErrRecordNotFound by returning nil, nil
	var todo model.Todo
	err := r.db.First(&todo, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &todo, nil
}

// FindAll retrieves all todos.
func (r *TodoRepository) FindAll() ([]model.Todo, error) {
	// TODO: implement using r.db.Find(&todos).Error
	var todos []model.Todo
	err := r.db.Find(&todos).Error
	if err != nil {
		return nil, err
	}
	return todos, nil
}

// Update modifies an existing todo.
func (r *TodoRepository) Update(todo *model.Todo) error {
	// TODO: implement using r.db.Save(todo).Error
	err := r.db.Save(todo).Error
	if err != nil {
		return err
	}
	return nil
}

// Delete removes a todo by ID.
func (r *TodoRepository) Delete(id uint) error {
	// TODO: implement using r.db.Delete(&model.Todo{}, id).Error
	err := r.db.Delete(&model.Todo{}, id).Error
	if err != nil {
		return err
	}
	return nil
}

// CreateAndMarkDone creates a todo and marks it done in a transaction.
func (r *TodoRepository) CreateAndMarkDone(todo *model.Todo) error {
	// TODO: optional challenge
	// Use r.db.Transaction(func(tx *gorm.DB) error { ... })
	err := r.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Create(todo).Error
		if err != nil {
			return err
		}
		todo.Done = true
		err = tx.Save(todo).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}