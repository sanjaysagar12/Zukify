package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"zukify.com/database"
	"zukify.com/handlers"
)

func main() {
	// Initialize database connections
	err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.UserDB.Close()
	defer database.WorkspaceDB.Close()

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Public routes
	e.POST("/register", handlers.HandlerPostRegister)
	e.POST("/login", handlers.HandlerPostLogin)

	// Protected routes
	r := e.Group("/api")
	r.Use(handlers.JWTMiddleware)
	r.GET("/verify", handlers.HandlerVerifyToken)
	r.POST("/workspace", handlers.HandlerCreateWorkspace) // New route for workspace creation
	r.POST("/workspace/saveat", handlers.HandlerSaveAT)
	r.POST("/workspace/saveflow", handlers.HandlerSaveFlow)
	e.POST("/runAT",handlers.HandlePostAT)
	// Start server
	e.Logger.Fatal(e.Start(":80"))
}