package middleware

import (
	"log"
	"strings"

	"github.com/CanYangTang/go_learning/internal/auth"
	"github.com/CanYangTang/go_learning/pkg/apperror"
	"github.com/CanYangTang/go_learning/pkg/response"
	"github.com/gin-gonic/gin"
)

// userIDContextKey is where RequireAuth stashes the authenticated user id so
// downstream handlers can read it via UserIDFromContext.
const userIDContextKey = "user_id"

// RequireAuth verifies the Authorization: Bearer <token> header. A missing,
// malformed, expired, or otherwise invalid token results in 401 via the shared
// error envelope; a valid token stores the user id in the context and continues.
func RequireAuth(m *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token, ok := bearerToken(header)
		if !ok {
			response.WriteError(c, apperror.Unauthorized("missing or malformed authorization header"))
			return
		}

		userID, err := m.Parse(token)
		if err != nil {
			// 原始错误（过期/签名错/格式错）只写日志，不回给客户端。
			log.Printf("request_id=%s auth_error=%v", RequestIDFromContext(c), err)
			response.WriteError(c, apperror.Unauthorized("invalid or expired token"))
			return
		}

		// 存进 context，未来按用户隔离时 handler 直接取（今天还没用到）。
		c.Set(userIDContextKey, userID)
		c.Next()
	}
}

// bearerToken extracts the token from a "Bearer <token>" header value.
// It returns ok=false when the header is missing or not in Bearer form.
func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	return strings.TrimSpace(header[len(prefix):]), true
}

// UserIDFromContext returns the authenticated user id set by RequireAuth.
// It is the convention entry point for later per-user data isolation.
func UserIDFromContext(c *gin.Context) (uint, bool) {
	v, exists := c.Get(userIDContextKey)
	if !exists {
		return 0, false
	}
	id, ok := v.(uint)
	return id, ok
}
