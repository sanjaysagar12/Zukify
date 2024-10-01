package main

import (

	"net/http"

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

	// CORS middleware configuration
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://zukify.portos.site"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	// Public routes
	e.POST("/register", handlers.HandlerPostRegister)
	e.POST("/login", handlers.HandlerPostLogin)

	// Protected routes
	r := e.Group("/api")
	r.Use(handlers.JWTMiddleware)
	r.GET("/verify", (handlers.HandlerVerifyToken))
	r.POST("/workspace", handlers.HandlerCreateWorkspace) // New route for workspace creation
	r.GET("/getworkspace", handlers.HandlerGetWorkspaces)
	r.POST("/workspace/saveat", handlers.HandlerSaveAT)
	r.POST("/workspace/saveflow", handlers.HandlerSaveFlow)
	r.GET("/workspace/fetchpathat", handlers.HandlerFetchPathAT)
	r.GET("/workspace/fetchallat", handlers.HandlerFetchAllAT)
	r.GET("/workspace/fetchpathflow", handlers.HandlerFetchPathFlow)
	r.GET("/workspace/fetchallflow", handlers.HandlerFetchAllFlow)
	r.POST("/collaborator", handlers.HandlerAddCollaborator)
	e.POST("/runAT", handlers.HandlePostAT)
	e.POST("/runATFromSaved",handlers.HandlePostATFromSaved)
	
	r.GET("/load-flow", handlers.LoadSpecificFlow)
	r.POST("/save-flow", handlers.SaveFlow)
	// r.GET("/runATsave", handlers.HandlerRunATSave)


	// Start server
	e.Logger.Fatal(e.Start(":80"))
}

