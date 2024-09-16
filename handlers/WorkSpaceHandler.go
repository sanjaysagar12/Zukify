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

func HandlerFetchPathAT(c echo.Context) error {
	user := c.Get("user").(jwt.MapClaims)
	uid, ok := user["uid"].(float64)
	if !ok {
		log.Printf("Failed to extract UID from token: %v", user)
		return echo.NewHTTPError(http.StatusInternalServerError, "Invalid token")
	}

	wid := c.QueryParam("wid")
	if wid == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Workspace ID (wid) is required")
	}

	// Check if user has access to the workspace
	hasAccess, err := database.UserHasAccessToWorkspace(int(uid), wid)
	if err != nil {
		log.Printf("Failed to check workspace access: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify workspace access")
	}
	if !hasAccess {
		return echo.NewHTTPError(http.StatusForbidden, "You don't have access to this workspace")
	}

	// Fetch path AT data
	pathATData, err := database.FetchPathAT(wid)
	if err != nil {
		log.Printf("Failed to fetch path AT data: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch path AT data")
	}

	return c.JSON(http.StatusOK, pathATData)
}

func HandlerFetchAllAT(c echo.Context) error {
	user := c.Get("user").(jwt.MapClaims)
	uid, ok := user["uid"].(float64)
	if !ok {
		log.Printf("Failed to extract UID from token: %v", user)
		return echo.NewHTTPError(http.StatusInternalServerError, "Invalid token")
	}

	wid := c.QueryParam("wid")
	if wid == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Workspace ID (wid) is required")
	}

	id := c.QueryParam("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID is required")
	}

	// Check if user has access to the workspace
	hasAccess, err := database.UserHasAccessToWorkspace(int(uid), wid)
	if err != nil {
		log.Printf("Failed to check workspace access: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify workspace access")
	}
	if !hasAccess {
		return echo.NewHTTPError(http.StatusForbidden, "You don't have access to this workspace")
	}

	// Fetch all AT data
	allATData, err := database.FetchAllAT(wid, id)
	if err != nil {
		log.Printf("Failed to fetch all AT data: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch all AT data")
	}

	return c.JSON(http.StatusOK, allATData)
}

func HandlerFetchPathFlow(c echo.Context) error {
	user := c.Get("user").(jwt.MapClaims)
	uid, ok := user["uid"].(float64)
	if !ok {
		log.Printf("Failed to extract UID from token: %v", user)
		return echo.NewHTTPError(http.StatusInternalServerError, "Invalid token")
	}

	wid := c.QueryParam("wid")
	if wid == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Workspace ID (wid) is required")
	}

	// Check if user has access to the workspace
	hasAccess, err := database.UserHasAccessToWorkspace(int(uid), wid)
	if err != nil {
		log.Printf("Failed to check workspace access: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify workspace access")
	}
	if !hasAccess {
		return echo.NewHTTPError(http.StatusForbidden, "You don't have access to this workspace")
	}

	// Fetch path Flow data
	pathFlowData, err := database.FetchPathFlow(wid)
	if err != nil {
		log.Printf("Failed to fetch path Flow data: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch path Flow data")
	}

	return c.JSON(http.StatusOK, pathFlowData)
}

func HandlerFetchAllFlow(c echo.Context) error {
	user := c.Get("user").(jwt.MapClaims)
	uid, ok := user["uid"].(float64)
	if !ok {
		log.Printf("Failed to extract UID from token: %v", user)
		return echo.NewHTTPError(http.StatusInternalServerError, "Invalid token")
	}

	wid := c.QueryParam("wid")
	if wid == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Workspace ID (wid) is required")
	}

	fid := c.QueryParam("fid")
	if fid == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Flow ID (fid) is required")
	}

	// Check if user has access to the workspace
	hasAccess, err := database.UserHasAccessToWorkspace(int(uid), wid)
	if err != nil {
		log.Printf("Failed to check workspace access: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify workspace access")
	}
	if !hasAccess {
		return echo.NewHTTPError(http.StatusForbidden, "You don't have access to this workspace")
	}

	// Fetch all Flow data
	allFlowData, err := database.FetchAllFlow(wid, fid)
	if err != nil {
		log.Printf("Failed to fetch all Flow data: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch all Flow data")
	}

	return c.JSON(http.StatusOK, allFlowData)
}

func HandlerGetWorkspaces(c echo.Context) error {
	user := c.Get("user").(jwt.MapClaims)
	uid, ok := user["uid"].(float64)
	if !ok {
		log.Printf("Failed to extract UID from token: %v", user)
		return echo.NewHTTPError(http.StatusInternalServerError, "Invalid token")
	}

	workspaces, err := database.GetUserWorkspaces(int(uid))
	if err != nil {
		log.Printf("Failed to get user workspaces: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve workspaces")
	}

	type WorkspaceResponse struct {
		WID  string `json:"wid"`
		Name string `json:"name"`
	}

	var response []WorkspaceResponse
	for _, ws := range workspaces {
		response = append(response, WorkspaceResponse{
			WID:  ws.WID,
			Name: ws.Name,
		})
	}

	return c.JSON(http.StatusOK, response)
}