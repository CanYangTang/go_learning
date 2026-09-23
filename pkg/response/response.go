package response

import (
	"errors"
	"net/http"

	"github.com/CanYangTang/go_learning/pkg/apperror"
	"github.com/gin-gonic/gin"
)

// Body is the success envelope: {"data": ..., "message": "ok"}. Data uses
// omitempty so a nil payload (e.g. DELETE) marshals to just {"message":"ok"}.
type Body struct {
	Data    any    `json:"data,omitempty"`
	Message string `json:"message"`
}

// ErrorBody is the error envelope: {"error": {"code": ..., "message": ...}}.
type ErrorBody struct {
	Error ErrorPayload `json:"error"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteError is the single exit for error responses, shared by handlers and
// middleware. It uses AbortWithStatusJSON so the same call works in a handler
// (writes status + envelope, then the handler returns) and in middleware
// (additionally stops the downstream chain). errors.As unwraps a known
// apperror.Error and uses its StatusCode/Code/Message; anything else becomes a
// 500 with a fixed message — the raw error must never reach the client.
func WriteError(c *gin.Context, err error) {
	var appErr apperror.Error
	if errors.As(err, &appErr) {
		c.AbortWithStatusJSON(appErr.StatusCode, ErrorBody{
			Error: ErrorPayload{Code: appErr.Code, Message: appErr.Message},
		})
		return
	}
	c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorBody{
		Error: ErrorPayload{Code: "INTERNAL_ERROR", Message: "internal server error"},
	})
}

// WriteSuccess is the single exit for success responses. message is always
// "ok"; a nil data is omitted by Body's omitempty. It uses c.JSON (not Abort)
// because a handler is the tail of the chain and has nothing to stop.
func WriteSuccess(c *gin.Context, statusCode int, data any) {
	c.JSON(statusCode, Body{Data: data, Message: "ok"})
}
