package handler

import (
	"log"
	"net/http"

	"github.com/CanYangTang/go_learning/internal/auth"
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

// LoginResponse is UserResponse plus the freshly minted JWT.
type LoginResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Token string `json:"token"`
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
	tokens  *auth.Manager
}

// NewUserHandler creates a new UserHandler. The token manager is used by Login
// to mint a JWT on success.
func NewUserHandler(service UserService, tokens *auth.Manager) *UserHandler {
	return &UserHandler{service: service, tokens: tokens}
}

// Register handles POST /api/v1/users/register.
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("request_id=%s bind_error=%v", middleware.RequestIDFromContext(c), err)
		response.WriteError(c, apperror.Validation("invalid request body"))
		return
	}

	user, err := h.service.Register(req.Email, req.Password)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.WriteSuccess(c, http.StatusCreated, UserResponse{ID: user.ID, Email: user.Email})
}

// Login handles POST /api/v1/users/login. On success it mints a JWT and
// returns it in the response body alongside the user's id and email.
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("request_id=%s bind_error=%v", middleware.RequestIDFromContext(c), err)
		response.WriteError(c, apperror.Validation("invalid request body"))
		return
	}

	user, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	token, err := h.tokens.Generate(user.ID)
	if err != nil {
		log.Printf("request_id=%s token_error=%v", middleware.RequestIDFromContext(c), err)
		response.WriteError(c, apperror.Internal("login failed"))
		return
	}

	response.WriteSuccess(c, http.StatusOK, LoginResponse{ID: user.ID, Email: user.Email, Token: token})
}
