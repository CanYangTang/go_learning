package main

import (
	"log"
	"net/http"

	"github.com/CanYangTang/go_learning/internal/handler"
	"github.com/gin-gonic/gin"
)

func main() {
	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)

	// Create a new Gin router
	router := gin.New()

	// Create API v1 route group
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", handler.HealthHandler)
	}
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "not found",
		})
	})

	// Start the server
	addr := ":8080"
	log.Printf("server listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatal(err)
	}

}
