package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logging prints a short summary for each request.
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		requestID := RequestIDFromContext(c)
		status := c.Writer.Status()
		latency := time.Since(start).Truncate(time.Millisecond)
		path := c.Request.URL.RequestURI()

		if requestID != "" {
			log.Printf("request_id=%s method=%s path=%s status=%d latency=%s", requestID, c.Request.Method, path, status, latency)
			return
		}

		log.Printf("method=%s path=%s status=%d latency=%s", c.Request.Method, path, status, latency)
	}
}
