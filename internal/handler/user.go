package handler

import (
	"log"
	"net/http"

	"github.com/CanYangTang/go_learning/internal/middleware"
	"github.com/CanYangTang/go_learning/internal/model"
	"github.com/CanYangTang/go_learning/pkg/apperror"
	"github.com/CanYangTang/go_learning/pkg/response"
	"github.com/gin-gonic/gin"
)

// UserService is the business contract the handler depends on.
// Same rule as TodoService: the consumer declares the interface.
type UserService interface {
	Register(email, password string) (*model.User, error)
	Login(email, password string) (*model.User, error)
}

// UserResponse is the JSON shape returned to clients. Defined separately
// from model.User (which already has json:"-" on PasswordHash) as a second,
// independent line of defense: if that tag is ever removed by accident,
// this struct still can't leak the hash because the field doesn't exist here.
type UserResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
}

// RegisterRequest represents the request body for POST /api/v1/users/register.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginRequest represents the request body for POST /api/v1/users/login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	service UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{service: service}
}

// Register handles POST /api/v1/users/register.
//
// TODO: implement, following CreateTodo's shape in todo.go:
//   - c.ShouldBindJSON(&req); on error, log.Printf with
//     middleware.RequestIDFromContext(c) and the raw err, then
//     writeError(c, apperror.Validation("invalid request body")) and return.
//   - Call h.service.Register(req.Email, req.Password); on error, writeError(c, err).
//   - On success, c.JSON(http.StatusCreated, response.Body{
//     Data: UserResponse{ID: user.ID, Email: user.Email}, Message: "ok"}).
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("request_id=%s bind_error=%v", middleware.RequestIDFromContext(c), err)
		writeError(c, apperror.Validation("invalid request body"))
		return
	}

	user, err := h.service.Register(req.Email, req.Password)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Body{
		Data:    UserResponse{ID: user.ID, Email: user.Email},
		Message: "ok",
	})
}

// Login handles POST /api/v1/users/login.
//
// TODO: implement, same shape as Register but calling h.service.Login and
// responding with http.StatusOK instead of http.StatusCreated. No token is
// issued yet - that lands in Day 25 with JWT.
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("request_id=%s bind_error=%v", middleware.RequestIDFromContext(c), err)
		writeError(c, apperror.Validation("invalid request body"))
		return
	}
	user, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Body{
		Data:    UserResponse{ID: user.ID, Email: user.Email},
		Message: "ok",
	})
}
