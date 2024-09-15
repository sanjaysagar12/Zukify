package handlers

import (
	"net/http"
	"crypto/sha256"
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/golang-jwt/jwt"
	"zukify.com/database"
)

func HandlerCreateWorkspace(c echo.Context) error {
	user := c.Get("user").(jwt.MapClaims)
	uid, ok := user["uid"].(float64)
	if !ok {
		log.Printf("Failed to extract UID from token: %v", user)
		return echo.NewHTTPError(http.StatusInternalServerError, "Invalid token")
	}

	var req struct {
		WorkspaceName string `json:"workspace_name"`
	}
	if err := c.Bind(&req); err != nil {
		log.Printf("Failed to bind request: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.WorkspaceName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Workspace name is required")
	}

	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%d%s", int(uid), req.WorkspaceName)))
	hashedValue := fmt.Sprintf("%x", hash.Sum(nil))
	tablePrefix := hashedValue[:16]

	// Check if workspace already exists
	exists, err := database.WorkspaceExists(int(uid), req.WorkspaceName)
	if err != nil {
		log.Printf("Failed to check workspace existence: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to check workspace existence")
	}
	if exists {
		return echo.NewHTTPError(http.StatusConflict, "Workspace with this name already exists")
	}

	// Create AT table
	if err := database.CreateATTable(tablePrefix); err != nil {
		log.Printf("Failed to create AT table: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create workspace")
	}

	// Create Data table
	if err := database.CreateDataTable(tablePrefix, int(uid)); err != nil {
		log.Printf("Failed to create Data table: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create workspace")
	}

	// Create Flow table
	if err := database.CreateFlowTable(tablePrefix); err != nil {
		log.Printf("Failed to create Flow table: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create workspace")
	}

	// Add workspace to user
	err = database.AddWorkspaceToUser(int(uid), tablePrefix, req.WorkspaceName)
	if err != nil {
		log.Printf("Failed to update user workspace: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user workspace")
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "Workspace created successfully",
		"wid": tablePrefix,
	})
}