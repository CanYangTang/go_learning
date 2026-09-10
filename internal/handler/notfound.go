package handler

import (
	"github.com/CanYangTang/go_learning/pkg/apperror"
	"github.com/gin-gonic/gin"
)

// NotFoundHandler is mounted with router.NoRoute so unknown paths return the
// shared error envelope instead of an ad-hoc shape.
//
// It lives in this package on purpose: writeError is unexported, so the only
// way to reuse the single error exit is to be inside internal/handler.
func NotFoundHandler(c *gin.Context) {
	// The message stays generic - echoing the requested path back would reflect
	// client-controlled input into the response.
	writeError(c, apperror.NotFound("route not found"))
}
