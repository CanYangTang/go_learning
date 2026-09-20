package handler

import (
	"io"
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
	getResult    *model.Todo
	getErr       error
	updateResult *model.Todo
	updateErr    error
	deleteErr    error
	createdTitle string
	createCalls  int
	getID        uint
	getCalls     int
	updateID     uint
	updateTitle  string
	updateDone   bool
	updateCalls  int
	deleteID     uint
	deleteCalls  int
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

func (f *fakeTodoService) GetTodo(id uint) (*model.Todo, error) {
	f.getCalls++
	f.getID = id
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.getResult, nil
}

func (f *fakeTodoService) UpdateTodo(id uint, title string, done bool) (*model.Todo, error) {
	f.updateCalls++
	f.updateID = id
	f.updateTitle = title
	f.updateDone = done
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	return f.updateResult, nil
}

func (f *fakeTodoService) DeleteTodo(id uint) error {
	f.deleteCalls++
	f.deleteID = id
	return f.deleteErr
}

func newTodoRouter(svc TodoService) *gin.Engine {
	gin.SetMode(gin.TestMode)

	h := NewTodoHandler(svc)
	router := gin.New()
	router.POST("/api/v1/todos", h.CreateTodo)
	router.GET("/api/v1/todos", h.ListTodos)
	router.GET("/api/v1/todos/:id", h.GetTodo)
	router.PUT("/api/v1/todos/:id", h.UpdateTodo)
	router.DELETE("/api/v1/todos/:id", h.DeleteTodo)
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

func doRequest(t *testing.T, router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
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

// The bind error must not travel to the client verbatim. Before this test the
// three cases below returned things like
//
//	"Key: 'CreateTodoRequest.Title' Error:Field validation for 'Title' failed..."
//	"json: cannot unmarshal array into Go value of type handler.CreateTodoRequest"
//
// which hands out internal Go type names for free and tells a legitimate client
// nothing it can act on.
//
// This asserts the absence of the leak rather than one exact wording, so the
// message text stays a free choice.
func TestCreateTodoDoesNotLeakBindErrorDetails(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"missing field", `{}`},
		{"truncated json", `{"title":`},
		{"wrong json type", `[]`},
	}

	leaks := []string{
		"CreateTodoRequest",
		"Field validation",
		"cannot unmarshal",
		"unexpected EOF",
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &fakeTodoService{}
			router := newTodoRouter(svc)

			recorder := postTodo(t, router, tc.body)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
			}
			assertBodyContains(t, recorder.Body.String(), `"code":"VALIDATION_ERROR"`)

			body := recorder.Body.String()
			for _, leak := range leaks {
				if strings.Contains(body, leak) {
					t.Fatalf("response body %q leaks internal detail %q", body, leak)
				}
			}

			if svc.createCalls != 0 {
				t.Fatalf("service should not be called when binding fails, got %d calls", svc.createCalls)
			}
		})
	}
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

func TestGetTodoByIDSuccess(t *testing.T) {
	svc := &fakeTodoService{getResult: &model.Todo{ID: 7, Title: "Learn Go", Done: true}}
	router := newTodoRouter(svc)

	recorder := doRequest(t, router, http.MethodGet, "/api/v1/todos/7", "")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	assertBodyContains(t, recorder.Body.String(), `"id":7`, `"title":"Learn Go"`, `"done":true`)
	if svc.getID != 7 {
		t.Fatalf("service received id %d, want 7", svc.getID)
	}
}

func TestGetTodoByIDInvalidID(t *testing.T) {
	svc := &fakeTodoService{}
	router := newTodoRouter(svc)

	recorder := doRequest(t, router, http.MethodGet, "/api/v1/todos/abc", "")

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	assertBodyContains(t, recorder.Body.String(), `"code":"VALIDATION_ERROR"`)
	if svc.getCalls != 0 {
		t.Fatalf("service should not be called for an invalid id, got %d calls", svc.getCalls)
	}
}

func TestGetTodoByIDNotFound(t *testing.T) {
	svc := &fakeTodoService{getErr: apperror.NotFound("todo not found")}
	router := newTodoRouter(svc)

	recorder := doRequest(t, router, http.MethodGet, "/api/v1/todos/999", "")

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
	assertBodyContains(t, recorder.Body.String(), `"code":"NOT_FOUND"`)
}

func TestUpdateTodoSuccess(t *testing.T) {
	svc := &fakeTodoService{updateResult: &model.Todo{ID: 3, Title: "new", Done: true}}
	router := newTodoRouter(svc)

	recorder := doRequest(t, router, http.MethodPut, "/api/v1/todos/3", `{"title":"new","done":true}`)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	assertBodyContains(t, recorder.Body.String(), `"id":3`, `"title":"new"`, `"done":true`)
	if svc.updateID != 3 || svc.updateTitle != "new" || !svc.updateDone {
		t.Fatalf("service received id=%d title=%q done=%v, want 3/new/true", svc.updateID, svc.updateTitle, svc.updateDone)
	}
}

func TestUpdateTodoAllowsDoneFalse(t *testing.T) {
	svc := &fakeTodoService{updateResult: &model.Todo{ID: 3, Title: "new", Done: false}}
	router := newTodoRouter(svc)

	recorder := doRequest(t, router, http.MethodPut, "/api/v1/todos/3", `{"title":"new","done":false}`)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (done:false must not be rejected)", recorder.Code, http.StatusOK)
	}
	if svc.updateCalls != 1 || svc.updateDone {
		t.Fatalf("service updateCalls=%d done=%v, want 1/false", svc.updateCalls, svc.updateDone)
	}
}

func TestUpdateTodoInvalidID(t *testing.T) {
	svc := &fakeTodoService{}
	router := newTodoRouter(svc)

	recorder := doRequest(t, router, http.MethodPut, "/api/v1/todos/abc", `{"title":"new"}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	assertBodyContains(t, recorder.Body.String(), `"code":"VALIDATION_ERROR"`)
	if svc.updateCalls != 0 {
		t.Fatalf("service should not be called for an invalid id, got %d calls", svc.updateCalls)
	}
}

func TestUpdateTodoInvalidJSON(t *testing.T) {
	svc := &fakeTodoService{}
	router := newTodoRouter(svc)

	recorder := doRequest(t, router, http.MethodPut, "/api/v1/todos/3", `{"title":invalid}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	assertBodyContains(t, recorder.Body.String(), `"code":"VALIDATION_ERROR"`)
	if svc.updateCalls != 0 {
		t.Fatalf("service should not be called when binding fails, got %d calls", svc.updateCalls)
	}
}

func TestUpdateTodoMissingTitle(t *testing.T) {
	svc := &fakeTodoService{}
	router := newTodoRouter(svc)

	recorder := doRequest(t, router, http.MethodPut, "/api/v1/todos/3", `{"done":true}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	assertBodyContains(t, recorder.Body.String(), `"code":"VALIDATION_ERROR"`)
	if svc.updateCalls != 0 {
		t.Fatalf("service should not be called when title is missing, got %d calls", svc.updateCalls)
	}
}

func TestUpdateTodoNotFound(t *testing.T) {
	svc := &fakeTodoService{updateErr: apperror.NotFound("todo not found")}
	router := newTodoRouter(svc)

	recorder := doRequest(t, router, http.MethodPut, "/api/v1/todos/999", `{"title":"new","done":true}`)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
	assertBodyContains(t, recorder.Body.String(), `"code":"NOT_FOUND"`)
}

func TestDeleteTodoSuccess(t *testing.T) {
	svc := &fakeTodoService{}
	router := newTodoRouter(svc)

	recorder := doRequest(t, router, http.MethodDelete, "/api/v1/todos/5", "")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	// response.Body.Data has json:"data,omitempty", so a nil Data is dropped
	// entirely: the delete success body is {"message":"ok"}, not {"data":null,...}.
	assertBodyContains(t, recorder.Body.String(), `"message":"ok"`)
	if strings.Contains(recorder.Body.String(), `"data"`) {
		t.Fatalf("expected no data field on delete success, got %q", recorder.Body.String())
	}
	if svc.deleteID != 5 {
		t.Fatalf("service received id %d, want 5", svc.deleteID)
	}
}

func TestDeleteTodoInvalidID(t *testing.T) {
	svc := &fakeTodoService{}
	router := newTodoRouter(svc)

	recorder := doRequest(t, router, http.MethodDelete, "/api/v1/todos/abc", "")

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	assertBodyContains(t, recorder.Body.String(), `"code":"VALIDATION_ERROR"`)
	if svc.deleteCalls != 0 {
		t.Fatalf("service should not be called for an invalid id, got %d calls", svc.deleteCalls)
	}
}

func TestDeleteTodoNotFound(t *testing.T) {
	svc := &fakeTodoService{deleteErr: apperror.NotFound("todo not found")}
	router := newTodoRouter(svc)

	recorder := doRequest(t, router, http.MethodDelete, "/api/v1/todos/999", "")

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
	assertBodyContains(t, recorder.Body.String(), `"code":"NOT_FOUND"`)
}
