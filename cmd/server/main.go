package main

import (
	"log"
	"os"

	"github.com/CanYangTang/go_learning/internal/config"
	"github.com/CanYangTang/go_learning/internal/handler"
	"github.com/CanYangTang/go_learning/internal/middleware"
	"github.com/CanYangTang/go_learning/internal/model"
	"github.com/CanYangTang/go_learning/internal/repository"
	"github.com/CanYangTang/go_learning/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	// Dependency injection happens here and nowhere else.
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = config.DefaultDSN()
	}

	db, err := config.ConnectGorm(dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Todo{}); err != nil {
		log.Fatal(err)
	}

	todoRepo := repository.NewTodoRepository(db)
	todoService := service.NewTodoService(todoRepo)
	todoHandler := handler.NewTodoHandler(todoService)

	router := gin.New()
	// Recovery sits after Logging so a recovered panic still produces an access
	// log line, and before CORS so the 500 keeps its cross-origin headers.
	router.Use(middleware.RequestID(), middleware.Logging(), middleware.Recovery(), middleware.CORS(), middleware.AuthPlaceholder())

	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", handler.HealthHandler)
		v1.POST("/todos", todoHandler.CreateTodo)
		v1.GET("/todos", todoHandler.ListTodos)
	}

	router.NoRoute(handler.NotFoundHandler)

	addr := ":8080"
	log.Printf("server listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatal(err)
	}
}
