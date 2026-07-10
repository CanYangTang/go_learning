package handler

import (
	"net/http"

	"github.com/CanYangTang/go_learning/pkg/response"
	"github.com/gin-gonic/gin"
)

// Todo represents a todo item.
type Todo struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// CreateTodoRequest represents the request body for creating a todo.
type CreateTodoRequest struct {
	Title string `json:"title" binding:"required"`
}

// TodoHandler handles todo-related HTTP requests.
type TodoHandler struct {
	// TODO: inject service or repository later
}

// NewTodoHandler creates a new TodoHandler.
func NewTodoHandler() *TodoHandler {
	return &TodoHandler{}
}

// CreateTodo handles POST /api/v1/todos.
func (h *TodoHandler) CreateTodo(c *gin.Context) {
	var req CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorBody{
			Error: response.ErrorPayload{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	todo := Todo{
		Title: req.Title,
		Done:  false,
	}

	c.JSON(http.StatusCreated, response.Body{
		Data:    todo,
		Message: "ok",
	})
}

// ListTodos handles GET /api/v1/todos.
func (h *TodoHandler) ListTodos(c *gin.Context) {
	todos := []Todo{}
	c.JSON(http.StatusOK, response.Body{
		Data:    todos,
		Message: "ok",
	})
}
