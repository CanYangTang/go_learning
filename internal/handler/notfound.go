package handler

import (
	"github.com/CanYangTang/go_learning/pkg/apperror"
	"github.com/CanYangTang/go_learning/pkg/response"
	"github.com/gin-gonic/gin"
)

// NotFoundHandler is mounted with router.NoRoute so unknown paths return the
// shared error envelope instead of an ad-hoc shape.
func NotFoundHandler(c *gin.Context) {
	// The message stays generic - echoing the requested path back would reflect
	// client-controlled input into the response.
	response.WriteError(c, apperror.NotFound("route not found"))
}
