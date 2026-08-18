package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	RequestIDHeader     = "X-Request-ID"
	RequestIDContextKey = "request_id"
)

// RequestID stores a request identifier on the Gin context and response header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader(RequestIDHeader))
		if requestID == "" {
			requestID = newRequestID()
		}

		c.Set(RequestIDContextKey, requestID)
		c.Writer.Header().Set(RequestIDHeader, requestID)
		c.Next()
	}
}

// RequestIDFromContext returns the request ID stored by RequestID().
func RequestIDFromContext(c *gin.Context) string {
	return c.GetString(RequestIDContextKey)
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fallbackRequestID()
	}
	return hex.EncodeToString(b[:])
}

func fallbackRequestID() string {
	return "request-id"
}
