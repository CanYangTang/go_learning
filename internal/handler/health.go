package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler returns the health status of the service.
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "go-learning",
		"version": "0.1.0",
	})
}
