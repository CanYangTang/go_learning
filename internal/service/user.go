package service

import (
	"errors"
	"strings"

	"github.com/CanYangTang/go_learning/internal/model"
	"github.com/CanYangTang/go_learning/pkg/apperror"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository is the persistence contract the service depends on.
// Same rule as TodoRepository: this interface lives in the consumer
// package (service), and only declares the methods the service uses.
type UserRepository interface {
	Create(user *model.User) error
	FindByEmail(email string) (*model.User, error)
}

// UserService holds the business rules for registration and login.
type UserService struct {
	repo UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Register validates the email and password, hashes the password with
// bcrypt, and stores a new user.
//
// TODO: implement
//   - Trim the email; reject empty email with apperror.Validation("email is required").
//   - Reject empty password with apperror.Validation("password is required").
//   - Look up the email via s.repo.FindByEmail; a lookup error becomes
//     apperror.Internal("register failed"); an existing user becomes
//     apperror.Validation("email already registered").
//   - Hash the password with bcrypt.GenerateFromPassword(..., bcrypt.DefaultCost).
//     bcrypt rejects passwords over 72 bytes - that error is a client input
//     problem, not a server fault: return apperror.Validation("password is too long"),
//     not apperror.Internal.
//   - Build a model.User{Email, PasswordHash: string(hash)} and call s.repo.Create.
//     A duplicate-key error surfaces here too (a concurrent request can win the
//     race between the FindByEmail check above and this Create) - the repository
//     layer is responsible for translating it into apperror.Validation, so this
//     layer only needs to wrap any other Create error as apperror.Internal("register failed").
func (s *UserService) Register(email, password string) (*model.User, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, apperror.Validation("email is required")
	}
	if len(password) == 0 {
		return nil, apperror.Validation("password is required")
	}

	existing, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, apperror.Internal("register failed")
	}
	if existing != nil {
		return nil, apperror.Validation("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		// This is bcrypt's 72-byte input limit, not a server fault.
		return nil, apperror.Validation("password is too long")
	}

	user := &model.User{Email: email, PasswordHash: string(hash)}
	if err := s.repo.Create(user); err != nil {
		// A concurrent request can win the race between the FindByEmail check
		// above and this Create: the repository already translates that into
		// apperror.Validation("email already registered"), so pass it through
		// as-is instead of flattening it into a 500.
		var appErr apperror.Error
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, apperror.Internal("register failed")
	}
	return user, nil
}

// Login validates the email and password and returns the matching user.
//
// TODO: implement
//   - Trim the email.
//   - Look up the user via s.repo.FindByEmail; a lookup error becomes
//     apperror.Internal("login failed").
//   - If no user is found, OR the password does not match
//     (bcrypt.CompareHashAndPassword returns a non-nil error), return the
//     SAME error: apperror.Validation("invalid email or password"). Do not
//     let a caller distinguish "no such email" from "wrong password" -
//     that distinction is account-enumeration information, not something
//     a legitimate client needs.
func (s *UserService) Login(email, password string) (*model.User, error) {
	email = strings.TrimSpace(email)
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, apperror.Internal("login failed")
	}
	if user == nil {
		return nil, apperror.Validation("invalid email or password")
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, apperror.Validation("invalid email or password")
	}
	return user, nil
}
