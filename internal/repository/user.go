package repository

import (
	"errors"

	"github.com/CanYangTang/go_learning/internal/model"
	"github.com/CanYangTang/go_learning/pkg/apperror"
	"gorm.io/gorm"
)

// UserRepository handles database operations for users using GORM.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user and sets the ID.
//
// TODO: implement
//   - Insert via r.db.Create(user).
//   - If the error satisfies errors.Is(err, gorm.ErrDuplicatedKey), return
//     apperror.Validation("email already registered") instead of the raw
//     *mysql.MySQLError - this only works if internal/config/gorm.go opens
//     the connection with gorm.Config{TranslateError: true}, so check that
//     setting is in place before assuming this branch is reachable.
//   - Any other error: return it as-is (the service layer wraps it).
func (r *UserRepository) Create(user *model.User) error {
	if err := r.db.Create(user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return apperror.Validation("email already registered")
		}
		return err
	}
	return nil
}

// FindByEmail retrieves a user by email. Returns (nil, nil) if not found -
// same convention as TodoRepository.FindByID.
//
// TODO: implement
//   - Query with r.db.Where("email = ?", email).First(&user).
//   - Translate gorm.ErrRecordNotFound into (nil, nil).
//   - Any other error: return (nil, err).
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
