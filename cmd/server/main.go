package main

import (
	"log"
	"os"
	"time"

	"github.com/CanYangTang/go_learning/internal/auth"
	"github.com/CanYangTang/go_learning/internal/config"
	"github.com/CanYangTang/go_learning/internal/handler"
	"github.com/CanYangTang/go_learning/internal/middleware"
	"github.com/CanYangTang/go_learning/internal/model"
	"github.com/CanYangTang/go_learning/internal/repository"
	"github.com/CanYangTang/go_learning/internal/service"
	"github.com/gin-gonic/gin"
)

// Compile-time proof that each concrete type satisfies the consumer-declared
// interface it is injected into below. These bindings are also enforced at the
// wiring calls, but stating them here keeps the layer contract from silently
// breaking if a signature drifts.
var (
	_ service.TodoRepository = (*repository.TodoRepository)(nil)
	_ service.UserRepository = (*repository.UserRepository)(nil)
	_ handler.TodoService    = (*service.TodoService)(nil)
	_ handler.UserService    = (*service.UserService)(nil)
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
	if err := db.AutoMigrate(&model.Todo{}, &model.User{}); err != nil {
		log.Fatal(err)
	}

	// The signing key comes from JWT_SECRET (dev default otherwise). Tokens are
	// valid for 24h. Never log the secret.
	tokenManager := auth.NewManager([]byte(config.JWTSecret()), 24*time.Hour)

	todoRepo := repository.NewTodoRepository(db)
	todoService := service.NewTodoService(todoRepo)
	todoHandler := handler.NewTodoHandler(todoService)

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService, tokenManager)

	router := gin.New()
	// Recovery sits after Logging so a recovered panic still produces an access
	// log line, and before CORS so the 500 keeps its cross-origin headers.
	// Auth is NOT global: it is mounted on the protected group below so that
	// /health, /users/register and /users/login stay public (backlog A6).
	router.Use(middleware.RequestID(), middleware.Logging(), middleware.Recovery(), middleware.CORS(config.AllowedOrigins()))

	v1 := router.Group("/api/v1")
	{
		// Public: no token required.
		v1.GET("/health", handler.HealthHandler)
		v1.POST("/users/register", userHandler.Register)
		v1.POST("/users/login", userHandler.Login)

		// Protected: a valid Bearer token is required.
		protected := v1.Group("")
		protected.Use(middleware.RequireAuth(tokenManager))
		{
			protected.POST("/todos", todoHandler.CreateTodo)
			protected.GET("/todos", todoHandler.ListTodos)
			protected.GET("/todos/:id", todoHandler.GetTodo)
			protected.PUT("/todos/:id", todoHandler.UpdateTodo)
			protected.DELETE("/todos/:id", todoHandler.DeleteTodo)
		}
	}

	router.NoRoute(handler.NotFoundHandler)

	addr := ":8080"
	log.Printf("server listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatal(err)
	}
}
