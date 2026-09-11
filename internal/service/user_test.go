package service

import (
	"errors"
	"net/http"
	"testing"

	"github.com/CanYangTang/go_learning/internal/model"
	"github.com/CanYangTang/go_learning/internal/repository"
	"github.com/CanYangTang/go_learning/pkg/apperror"
	"golang.org/x/crypto/bcrypt"
)

// The real repository must satisfy the interface declared by the service.
var _ UserRepository = (*repository.UserRepository)(nil)

// fakeUserRepo is an in-memory stand-in for the database.
type fakeUserRepo struct {
	users          []model.User
	nextID         uint
	createErr      error
	findByEmailErr error
	createCalls    int
	findEmailCalls int
}

var _ UserRepository = (*fakeUserRepo)(nil)

func (f *fakeUserRepo) Create(user *model.User) error {
	f.createCalls++
	if f.createErr != nil {
		return f.createErr
	}
	f.nextID++
	user.ID = f.nextID
	f.users = append(f.users, *user)
	return nil
}

func (f *fakeUserRepo) FindByEmail(email string) (*model.User, error) {
	f.findEmailCalls++
	if f.findByEmailErr != nil {
		return nil, f.findByEmailErr
	}
	for _, u := range f.users {
		if u.Email == email {
			return &u, nil
		}
	}
	return nil, nil
}

func assertUserAppError(t *testing.T, err error, wantCode string, wantStatus int) {
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

func TestRegisterStoresUserWithHashedPassword(t *testing.T) {
	repo := &fakeUserRepo{}
	svc := NewUserService(repo)

	user, err := svc.Register("a@example.com", "hunter2")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if user == nil {
		t.Fatalf("expected a user, got nil")
	}
	if user.Email != "a@example.com" {
		t.Fatalf("email = %q, want %q", user.Email, "a@example.com")
	}
	if user.PasswordHash == "" || user.PasswordHash == "hunter2" {
		t.Fatalf("password should be hashed, got %q", user.PasswordHash)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("hunter2")); err != nil {
		t.Fatalf("stored hash does not match the original password: %v", err)
	}
}

func TestRegisterRejectsEmptyEmail(t *testing.T) {
	repo := &fakeUserRepo{}
	svc := NewUserService(repo)

	user, err := svc.Register("", "hunter2")

	assertUserAppError(t, err, "VALIDATION_ERROR", http.StatusBadRequest)
	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}
}

func TestRegisterRejectsEmptyPassword(t *testing.T) {
	repo := &fakeUserRepo{}
	svc := NewUserService(repo)

	user, err := svc.Register("a@example.com", "")

	assertUserAppError(t, err, "VALIDATION_ERROR", http.StatusBadRequest)
	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}
}

func TestRegisterRejectsDuplicateEmail(t *testing.T) {
	repo := &fakeUserRepo{}
	svc := NewUserService(repo)

	if _, err := svc.Register("a@example.com", "hunter2"); err != nil {
		t.Fatalf("first Register failed: %v", err)
	}

	user, err := svc.Register("a@example.com", "other-password")

	assertUserAppError(t, err, "VALIDATION_ERROR", http.StatusBadRequest)
	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}
	if repo.createCalls != 1 {
		t.Fatalf("repo.Create should not be called for a duplicate email, got %d calls", repo.createCalls)
	}
}

func TestRegisterRejectsPasswordOverBcryptLimit(t *testing.T) {
	repo := &fakeUserRepo{}
	svc := NewUserService(repo)

	longPassword := make([]byte, 80)
	for i := range longPassword {
		longPassword[i] = 'a'
	}

	user, err := svc.Register("a@example.com", string(longPassword))

	assertUserAppError(t, err, "VALIDATION_ERROR", http.StatusBadRequest)
	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}
}

func TestRegisterWrapsFindByEmailError(t *testing.T) {
	repo := &fakeUserRepo{findByEmailErr: errors.New("db is down")}
	svc := NewUserService(repo)

	user, err := svc.Register("a@example.com", "hunter2")

	assertUserAppError(t, err, "INTERNAL_ERROR", http.StatusInternalServerError)
	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}
}

func TestRegisterWrapsCreateError(t *testing.T) {
	repo := &fakeUserRepo{createErr: errors.New("db is down")}
	svc := NewUserService(repo)

	user, err := svc.Register("a@example.com", "hunter2")

	assertUserAppError(t, err, "INTERNAL_ERROR", http.StatusInternalServerError)
	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}
}

func TestRegisterPropagatesRepositoryValidationError(t *testing.T) {
	// Simulates the concurrent-request race: FindByEmail found nothing, but
	// Create hits the database's unique index and the repository has already
	// translated that into apperror.Validation.
	repo := &fakeUserRepo{createErr: apperror.Validation("email already registered")}
	svc := NewUserService(repo)

	user, err := svc.Register("a@example.com", "hunter2")

	assertUserAppError(t, err, "VALIDATION_ERROR", http.StatusBadRequest)
	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}
}

func TestLoginSucceedsWithCorrectPassword(t *testing.T) {
	repo := &fakeUserRepo{}
	svc := NewUserService(repo)

	if _, err := svc.Register("a@example.com", "hunter2"); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	user, err := svc.Login("a@example.com", "hunter2")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if user == nil || user.Email != "a@example.com" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestLoginRejectsWrongPasswordAndUnknownEmailIdentically(t *testing.T) {
	repo := &fakeUserRepo{}
	svc := NewUserService(repo)

	if _, err := svc.Register("a@example.com", "hunter2"); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	_, wrongPasswordErr := svc.Login("a@example.com", "wrong-password")
	_, unknownEmailErr := svc.Login("nobody@example.com", "hunter2")

	var wrongPasswordAppErr, unknownEmailAppErr apperror.Error
	if !errors.As(wrongPasswordErr, &wrongPasswordAppErr) {
		t.Fatalf("wrong password error is not an apperror.Error: %v", wrongPasswordErr)
	}
	if !errors.As(unknownEmailErr, &unknownEmailAppErr) {
		t.Fatalf("unknown email error is not an apperror.Error: %v", unknownEmailErr)
	}

	if wrongPasswordAppErr.Code != unknownEmailAppErr.Code ||
		wrongPasswordAppErr.Message != unknownEmailAppErr.Message ||
		wrongPasswordAppErr.StatusCode != unknownEmailAppErr.StatusCode {
		t.Fatalf("wrong password and unknown email must return identical errors, got %+v vs %+v",
			wrongPasswordAppErr, unknownEmailAppErr)
	}
}

func TestLoginWrapsFindByEmailError(t *testing.T) {
	repo := &fakeUserRepo{findByEmailErr: errors.New("db is down")}
	svc := NewUserService(repo)

	user, err := svc.Login("a@example.com", "hunter2")

	assertUserAppError(t, err, "INTERNAL_ERROR", http.StatusInternalServerError)
	if user != nil {
		t.Fatalf("expected nil user, got %+v", user)
	}
}
