package handlers

import (
	"net/http"
	"crypto/sha256"
	"fmt"
	"log"
	// "encoding/json"

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

func HandlerSaveAT(c echo.Context) error {
	user := c.Get("user").(jwt.MapClaims)
	uid, ok := user["uid"].(float64)
	if !ok {
		log.Printf("Failed to extract UID from token: %v", user)
		return echo.NewHTTPError(http.StatusInternalServerError, "Invalid token")
	}

	var req struct {
		WorkspaceName string         `json:"workspace_name"`
		ATData        database.ATData `json:"at_data"`
	}
	if err := c.Bind(&req); err != nil {
		log.Printf("Failed to bind request: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.WorkspaceName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Workspace name is required")
	}

	// Check if user has access to the workspace
	workspaces, err := database.GetUserWorkspaces(int(uid))
	if err != nil {
		log.Printf("Failed to get user workspaces: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify workspace access")
	}

	var tablePrefix string
	for _, ws := range workspaces {
		if ws.Name == req.WorkspaceName {
			tablePrefix = ws.WID
			break
		}
	}

	if tablePrefix == "" {
		return echo.NewHTTPError(http.StatusForbidden, "You don't have access to this workspace")
	}

	// Save AT data
	err = database.SaveATData(tablePrefix, &req.ATData, int(uid))
	if err != nil {
		log.Printf("Failed to save AT data: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save AT data")
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "AT data saved successfully",
	})
}

func HandlerSaveFlow(c echo.Context) error {
	user := c.Get("user").(jwt.MapClaims)
	uid, ok := user["uid"].(float64)
	if !ok {
		log.Printf("Failed to extract UID from token: %v", user)
		return echo.NewHTTPError(http.StatusInternalServerError, "Invalid token")
	}

	var req struct {
		WorkspaceName string           `json:"workspace_name"`
		FlowData      database.FlowData `json:"flow_data"`
	}
	if err := c.Bind(&req); err != nil {
		log.Printf("Failed to bind request: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.WorkspaceName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Workspace name is required")
	}

	// Check if user has access to the workspace
	workspaces, err := database.GetUserWorkspaces(int(uid))
	if err != nil {
		log.Printf("Failed to get user workspaces: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify workspace access")
	}

	var tablePrefix string
	for _, ws := range workspaces {
		if ws.Name == req.WorkspaceName {
			tablePrefix = ws.WID
			break
		}
	}

	if tablePrefix == "" {
		return echo.NewHTTPError(http.StatusForbidden, "You don't have access to this workspace")
	}

	// Save Flow data
	err = database.SaveFlowData(tablePrefix, &req.FlowData, int(uid))
	if err != nil {
		log.Printf("Failed to save Flow data: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save Flow data")
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "Flow data saved successfully",
	})
}