package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CanYangTang/go_learning/internal/model"
	"github.com/CanYangTang/go_learning/pkg/apperror"
	"github.com/gin-gonic/gin"
)

// fakeUserService stands in for the real service layer, so handler tests
// need neither business rules, bcrypt, nor a database.
type fakeUserService struct {
	registerResult *model.User
	registerErr    error
	loginResult    *model.User
	loginErr       error
	registerCalls  int
	loginCalls     int
	lastEmail      string
	lastPassword   string
}

var _ UserService = (*fakeUserService)(nil)

func (f *fakeUserService) Register(email, password string) (*model.User, error) {
	f.registerCalls++
	f.lastEmail, f.lastPassword = email, password
	if f.registerErr != nil {
		return nil, f.registerErr
	}
	return f.registerResult, nil
}

func (f *fakeUserService) Login(email, password string) (*model.User, error) {
	f.loginCalls++
	f.lastEmail, f.lastPassword = email, password
	if f.loginErr != nil {
		return nil, f.loginErr
	}
	return f.loginResult, nil
}

func newUserRouter(svc UserService) *gin.Engine {
	gin.SetMode(gin.TestMode)

	h := NewUserHandler(svc)
	router := gin.New()
	router.POST("/api/v1/users/register", h.Register)
	router.POST("/api/v1/users/login", h.Login)
	return router
}

func postJSON(t *testing.T, router *gin.Engine, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func TestRegisterSuccess(t *testing.T) {
	svc := &fakeUserService{
		registerResult: &model.User{ID: 3, Email: "a@example.com", PasswordHash: "$2a$10$shouldneverleak"},
	}
	router := newUserRouter(svc)

	recorder := postJSON(t, router, "/api/v1/users/register", `{"email":"a@example.com","password":"hunter2"}`)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	assertBodyContains(t, recorder.Body.String(), `"id":3`, `"email":"a@example.com"`)
	if strings.Contains(recorder.Body.String(), "shouldneverleak") {
		t.Fatalf("response leaked the password hash: %s", recorder.Body.String())
	}
	if svc.registerCalls != 1 {
		t.Fatalf("service.Register called %d times, want 1", svc.registerCalls)
	}
}

func TestRegisterInvalidJSON(t *testing.T) {
	svc := &fakeUserService{}
	router := newUserRouter(svc)

	recorder := postJSON(t, router, "/api/v1/users/register", `{"email":invalid}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	assertBodyContains(t, recorder.Body.String(), `"code":"VALIDATION_ERROR"`)
	if svc.registerCalls != 0 {
		t.Fatalf("service should not be called when binding fails, got %d calls", svc.registerCalls)
	}
}

func TestRegisterMissingFields(t *testing.T) {
	svc := &fakeUserService{}
	router := newUserRouter(svc)

	recorder := postJSON(t, router, "/api/v1/users/register", `{}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	assertBodyContains(t, recorder.Body.String(), `"code":"VALIDATION_ERROR"`)
}

func TestRegisterDoesNotLeakBindErrorDetails(t *testing.T) {
	svc := &fakeUserService{}
	router := newUserRouter(svc)

	recorder := postJSON(t, router, "/api/v1/users/register", `{"email":`)

	body := recorder.Body.String()
	leaks := []string{"RegisterRequest", "Field validation", "cannot unmarshal", "unexpected EOF"}
	for _, leak := range leaks {
		if strings.Contains(body, leak) {
			t.Fatalf("response body %q leaks internal detail %q", body, leak)
		}
	}
}

func TestRegisterMapsServiceValidationError(t *testing.T) {
	svc := &fakeUserService{registerErr: apperror.Validation("email already registered")}
	router := newUserRouter(svc)

	recorder := postJSON(t, router, "/api/v1/users/register", `{"email":"a@example.com","password":"hunter2"}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	assertBodyContains(t, recorder.Body.String(), `"code":"VALIDATION_ERROR"`, `"message":"email already registered"`)
}

func TestRegisterMapsServiceInternalError(t *testing.T) {
	svc := &fakeUserService{registerErr: apperror.Internal("register failed")}
	router := newUserRouter(svc)

	recorder := postJSON(t, router, "/api/v1/users/register", `{"email":"a@example.com","password":"hunter2"}`)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	assertBodyContains(t, recorder.Body.String(), `"code":"INTERNAL_ERROR"`)
}

func TestLoginSuccess(t *testing.T) {
	svc := &fakeUserService{
		loginResult: &model.User{ID: 3, Email: "a@example.com", PasswordHash: "$2a$10$shouldneverleak"},
	}
	router := newUserRouter(svc)

	recorder := postJSON(t, router, "/api/v1/users/login", `{"email":"a@example.com","password":"hunter2"}`)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	assertBodyContains(t, recorder.Body.String(), `"id":3`, `"email":"a@example.com"`)
	if strings.Contains(recorder.Body.String(), "shouldneverleak") {
		t.Fatalf("response leaked the password hash: %s", recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "token") {
		t.Fatalf("no token should be issued until Day 25 (JWT): %s", recorder.Body.String())
	}
}

func TestLoginInvalidJSON(t *testing.T) {
	svc := &fakeUserService{}
	router := newUserRouter(svc)

	recorder := postJSON(t, router, "/api/v1/users/login", `{}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	assertBodyContains(t, recorder.Body.String(), `"code":"VALIDATION_ERROR"`)
	if svc.loginCalls != 0 {
		t.Fatalf("service should not be called when binding fails, got %d calls", svc.loginCalls)
	}
}

func TestLoginMapsServiceValidationError(t *testing.T) {
	svc := &fakeUserService{loginErr: apperror.Validation("invalid email or password")}
	router := newUserRouter(svc)

	recorder := postJSON(t, router, "/api/v1/users/login", `{"email":"a@example.com","password":"wrong"}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	assertBodyContains(t, recorder.Body.String(), `"code":"VALIDATION_ERROR"`, `"message":"invalid email or password"`)
}

func TestLoginMapsServiceInternalError(t *testing.T) {
	svc := &fakeUserService{loginErr: apperror.Internal("login failed")}
	router := newUserRouter(svc)

	recorder := postJSON(t, router, "/api/v1/users/login", `{"email":"a@example.com","password":"hunter2"}`)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	assertBodyContains(t, recorder.Body.String(), `"code":"INTERNAL_ERROR"`)
}
