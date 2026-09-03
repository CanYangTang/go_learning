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

// fakeTodoService stands in for the real service layer, so handler tests
// need neither business rules nor a database.
type fakeTodoService struct {
	createResult *model.Todo
	createErr    error
	listResult   []model.Todo
	listErr      error
	createdTitle string
	createCalls  int
}

var _ TodoService = (*fakeTodoService)(nil)

func (f *fakeTodoService) CreateTodo(title string) (*model.Todo, error) {
	f.createCalls++
	f.createdTitle = title
	if f.createErr != nil {
		return nil, f.createErr
	}
	return f.createResult, nil
}

func (f *fakeTodoService) ListTodos() ([]model.Todo, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listResult, nil
}

func newTodoRouter(svc TodoService) *gin.Engine {
	gin.SetMode(gin.TestMode)

	h := NewTodoHandler(svc)
	router := gin.New()
	router.POST("/api/v1/todos", h.CreateTodo)
	router.GET("/api/v1/todos", h.ListTodos)
	return router
}

func postTodo(t *testing.T, router *gin.Engine, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func getTodos(t *testing.T, router *gin.Engine) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/todos", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func assertBodyContains(t *testing.T, body string, want ...string) {
	t.Helper()

	for _, fragment := range want {
		if !strings.Contains(body, fragment) {
			t.Fatalf("response body %q does not contain %q", body, fragment)
		}
	}
}

func TestCreateTodoSuccess(t *testing.T) {
	svc := &fakeTodoService{
		createResult: &model.Todo{ID: 7, Title: "Learn Go", Done: false},
	}
	router := newTodoRouter(svc)

	recorder := postTodo(t, router, `{"title":"Learn Go"}`)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	assertBodyContains(t, recorder.Body.String(), `"id":7`, `"title":"Learn Go"`, `"done":false`)
	if svc.createCalls != 1 {
		t.Fatalf("service.CreateTodo called %d times, want 1", svc.createCalls)
	}
}

func TestCreateTodoPassesTitleThroughUntrimmed(t *testing.T) {
	svc := &fakeTodoService{
		createResult: &model.Todo{ID: 1, Title: "Learn Go"},
	}
	router := newTodoRouter(svc)

	postTodo(t, router, `{"title":"  Learn Go  "}`)

	if svc.createdTitle != "  Learn Go  " {
		t.Fatalf("service received %q, want the raw title (trimming belongs to the service)", svc.createdTitle)
	}
}

func TestCreateTodoInvalidJSON(t *testing.T) {
	svc := &fakeTodoService{}
	router := newTodoRouter(svc)

	recorder := postTodo(t, router, `{"title":invalid}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	assertBodyContains(t, recorder.Body.String(), `"code":"VALIDATION_ERROR"`)
	if svc.createCalls != 0 {
		t.Fatalf("service should not be called when binding fails, got %d calls", svc.createCalls)
	}
}

func TestCreateTodoMissingTitle(t *testing.T) {
	svc := &fakeTodoService{}
	router := newTodoRouter(svc)

	recorder := postTodo(t, router, `{}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	assertBodyContains(t, recorder.Body.String(), `"code":"VALIDATION_ERROR"`)
}

func TestCreateTodoMapsServiceValidationError(t *testing.T) {
	svc := &fakeTodoService{createErr: apperror.Validation("title is required")}
	router := newTodoRouter(svc)

	recorder := postTodo(t, router, `{"title":"   "}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	assertBodyContains(t, recorder.Body.String(), `"code":"VALIDATION_ERROR"`, `"message":"title is required"`)
}

func TestCreateTodoMapsServiceInternalError(t *testing.T) {
	svc := &fakeTodoService{createErr: apperror.Internal("create todo failed")}
	router := newTodoRouter(svc)

	recorder := postTodo(t, router, `{"title":"Learn Go"}`)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	assertBodyContains(t, recorder.Body.String(), `"code":"INTERNAL_ERROR"`)
}

func TestListTodosReturnsItems(t *testing.T) {
	svc := &fakeTodoService{
		listResult: []model.Todo{
			{ID: 1, Title: "Todo 1", Done: false},
			{ID: 2, Title: "Todo 2", Done: true},
		},
	}
	router := newTodoRouter(svc)

	recorder := getTodos(t, router)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	assertBodyContains(t, recorder.Body.String(),
		`"id":1`, `"title":"Todo 1"`, `"done":false`,
		`"id":2`, `"title":"Todo 2"`, `"done":true`,
	)
}

func TestListTodosEmpty(t *testing.T) {
	svc := &fakeTodoService{listResult: nil}
	router := newTodoRouter(svc)

	recorder := getTodos(t, router)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	assertBodyContains(t, recorder.Body.String(), `"data":[]`)
}

func TestListTodosMapsServiceError(t *testing.T) {
	svc := &fakeTodoService{listErr: apperror.Internal("list todos failed")}
	router := newTodoRouter(svc)

	recorder := getTodos(t, router)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	assertBodyContains(t, recorder.Body.String(), `"code":"INTERNAL_ERROR"`)
}
