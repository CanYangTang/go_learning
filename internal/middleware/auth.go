package middleware

import "github.com/gin-gonic/gin"

// AuthPlaceholder reserves a future JWT auth hook without blocking public routes.
func AuthPlaceholder() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
