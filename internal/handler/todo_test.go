package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCreateTodoSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewTodoHandler()
	router := gin.New()
	router.POST("/api/v1/todos", h.CreateTodo)

	body := `{"title":"Learn Go"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}

	// Check response contains expected fields
	respBody := recorder.Body.String()
	if !strings.Contains(respBody, `"title":"Learn Go"`) {
		t.Fatalf("response body %q does not contain title", respBody)
	}
	if !strings.Contains(respBody, `"done":false`) {
		t.Fatalf("response body %q does not contain done", respBody)
	}
}

func TestCreateTodoInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewTodoHandler()
	router := gin.New()
	router.POST("/api/v1/todos", h.CreateTodo)

	body := `{"title":invalid}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestCreateTodoMissingTitle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewTodoHandler()
	router := gin.New()
	router.POST("/api/v1/todos", h.CreateTodo)

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestListTodosEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewTodoHandler()
	router := gin.New()
	router.GET("/api/v1/todos", h.ListTodos)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/todos", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	// Check response is an empty array
	respBody := recorder.Body.String()
	if !strings.Contains(respBody, `"data":[]`) {
		t.Fatalf("response body %q should contain empty data array", respBody)
	}
}
