package main

import (
	"log"
	"net/http"

	"github.com/CanYangTang/go_learning/internal/handler"
	"github.com/CanYangTang/go_learning/internal/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(middleware.RequestID(), middleware.Logging(), middleware.CORS(), middleware.AuthPlaceholder())

	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", handler.HealthHandler)
	}

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "not found",
		})
	})

	addr := ":8080"
	log.Printf("server listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatal(err)
	}
}
