package repository

import (
	"errors"
	"testing"

	"github.com/CanYangTang/go_learning/internal/model"
	"github.com/CanYangTang/go_learning/pkg/apperror"
	"gorm.io/gorm"
)

func cleanupUsers(t *testing.T) {
	t.Helper()
	if testDB == nil {
		return
	}
	if err := testDB.Exec("DELETE FROM users").Error; err != nil {
		t.Fatalf("cleanup users failed: %v", err)
	}
}

func TestUserRepository_Create(t *testing.T) {
	if testDB == nil {
		t.Skip("integration tests require database connection")
	}
	cleanupUsers(t)

	repo := NewUserRepository(testDB)

	user := &model.User{Email: "a@example.com", PasswordHash: "hash"}

	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if user.ID == 0 {
		t.Fatalf("user.ID should be set after Create")
	}
}

func TestUserRepository_CreateDuplicateEmail(t *testing.T) {
	if testDB == nil {
		t.Skip("integration tests require database connection")
	}
	cleanupUsers(t)

	repo := NewUserRepository(testDB)

	first := &model.User{Email: "dup@example.com", PasswordHash: "hash1"}
	if err := repo.Create(first); err != nil {
		t.Fatalf("first Create failed: %v", err)
	}

	second := &model.User{Email: "dup@example.com", PasswordHash: "hash2"}
	err := repo.Create(second)

	if err == nil {
		t.Fatalf("expected an error for a duplicate email, got nil")
	}

	var appErr apperror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected an apperror.Error, got %v (%T) - this fails if "+
			"gorm.Config{TranslateError: true} is not set in internal/config/gorm.go, "+
			"since errors.Is(err, gorm.ErrDuplicatedKey) would never match", err, err)
	}
	if appErr.Code != "VALIDATION_ERROR" {
		t.Fatalf("code = %q, want %q", appErr.Code, "VALIDATION_ERROR")
	}
}

func TestUserRepository_FindByEmail(t *testing.T) {
	if testDB == nil {
		t.Skip("integration tests require database connection")
	}
	cleanupUsers(t)

	repo := NewUserRepository(testDB)

	user := &model.User{Email: "find@example.com", PasswordHash: "hash"}
	if err := repo.Create(user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByEmail("find@example.com")
	if err != nil {
		t.Fatalf("FindByEmail failed: %v", err)
	}
	if found == nil {
		t.Fatalf("expected to find user with email %q", "find@example.com")
	}
	if found.ID != user.ID {
		t.Fatalf("id = %d, want %d", found.ID, user.ID)
	}
}

func TestUserRepository_FindByEmailNotFound(t *testing.T) {
	if testDB == nil {
		t.Skip("integration tests require database connection")
	}
	cleanupUsers(t)

	repo := NewUserRepository(testDB)

	found, err := repo.FindByEmail("nobody@example.com")
	if err != nil {
		t.Fatalf("FindByEmail failed: %v", err)
	}
	if found != nil {
		t.Fatalf("expected nil user, got %+v", found)
	}
}

// Sanity check that TranslateError is actually wired up in the shared testDB
// connection - if this ever regresses, TestUserRepository_CreateDuplicateEmail
// above will fail with a clearer message, but this pins the underlying cause.
var _ = gorm.ErrDuplicatedKey
