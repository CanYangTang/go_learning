package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/CanYangTang/go_learning/internal/middleware"
	"github.com/CanYangTang/go_learning/internal/model"
	"github.com/CanYangTang/go_learning/pkg/apperror"
	"github.com/CanYangTang/go_learning/pkg/response"
	"github.com/gin-gonic/gin"
)

// TodoService is the business contract the handler depends on.
// Same rule as in the service package: the consumer declares the interface.
type TodoService interface {
	CreateTodo(title string) (*model.Todo, error)
	ListTodos() ([]model.Todo, error)
	GetTodo(id uint) (*model.Todo, error)
	UpdateTodo(id uint, title string, done bool) (*model.Todo, error)
	DeleteTodo(id uint) error
}

// Todo is the JSON shape returned to clients.
type Todo struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// CreateTodoRequest represents the request body for creating a todo.
type CreateTodoRequest struct {
	Title string `json:"title" binding:"required"`
}

// UpdateTodoRequest represents the request body for a full update (PUT).
// Done deliberately has no binding:"required": for a bool, `required` treats
// the legitimate value false as "missing" and rejects it. false is a valid
// state for Done, so we leave it off and let it default to the zero value.
type UpdateTodoRequest struct {
	Title string `json:"title" binding:"required"`
	Done  bool   `json:"done"`
}

// TodoHandler handles todo-related HTTP requests.
type TodoHandler struct {
	service TodoService
}

// NewTodoHandler creates a new TodoHandler.
func NewTodoHandler(service TodoService) *TodoHandler {
	return &TodoHandler{service: service}
}

// CreateTodo handles POST /api/v1/todos.
func (h *TodoHandler) CreateTodo(c *gin.Context) {
	var req CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// err.Error() is the raw validator / json error: it names internal types
		// such as CreateTodoRequest and tells a legitimate client nothing it can
		// act on. Keep it server-side, keyed by request ID so a client-reported
		// 400 can still be traced, and send a fixed message instead.
		log.Printf("request_id=%s bind_error=%v", middleware.RequestIDFromContext(c), err)
		response.WriteError(c, apperror.Validation("invalid request body: title is required"))
		return
	}

	// Pass the raw title through - trimming is the service's job.
	todo, err := h.service.CreateTodo(req.Title)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.WriteSuccess(c, http.StatusCreated, newTodoResponse(*todo))
}

// ListTodos handles GET /api/v1/todos.
func (h *TodoHandler) ListTodos(c *gin.Context) {
	todos, err := h.service.ListTodos()
	if err != nil {
		response.WriteError(c, err)
		return
	}

	// Start from an empty slice so an empty list marshals to [] instead of null.
	items := make([]Todo, 0, len(todos))
	for _, todo := range todos {
		items = append(items, newTodoResponse(todo))
	}

	response.WriteSuccess(c, http.StatusOK, items)
}

// newTodoResponse converts a domain model into the API response shape.
func newTodoResponse(todo model.Todo) Todo {
	return Todo{
		ID:    todo.ID,
		Title: todo.Title,
		Done:  todo.Done,
	}
}

// parseID converts a :id path segment into a uint.
//
// Parsing lives in the handler because "how an id is expressed in a URL" is an
// HTTP detail, not a business rule; the service only ever sees a clean uint.
func parseID(s string) (uint, error) {
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// GetTodo handles GET /api/v1/todos/:id.
func (h *TodoHandler) GetTodo(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.WriteError(c, apperror.Validation("invalid id"))
		return
	}

	todo, err := h.service.GetTodo(id)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.WriteSuccess(c, http.StatusOK, newTodoResponse(*todo))
}

// UpdateTodo handles PUT /api/v1/todos/:id.
func (h *TodoHandler) UpdateTodo(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.WriteError(c, apperror.Validation("invalid id"))
		return
	}

	var req UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("request_id=%s bind_error=%v", middleware.RequestIDFromContext(c), err)
		response.WriteError(c, apperror.Validation("invalid request body"))
		return
	}

	todo, err := h.service.UpdateTodo(id, req.Title, req.Done)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.WriteSuccess(c, http.StatusOK, newTodoResponse(*todo))
}

// DeleteTodo handles DELETE /api/v1/todos/:id. Success is 200 + envelope (not
// 204), to stay consistent with every other success response.
func (h *TodoHandler) DeleteTodo(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.WriteError(c, apperror.Validation("invalid id"))
		return
	}

	if err := h.service.DeleteTodo(id); err != nil {
		response.WriteError(c, err)
		return
	}

	response.WriteSuccess(c, http.StatusOK, nil)
}
