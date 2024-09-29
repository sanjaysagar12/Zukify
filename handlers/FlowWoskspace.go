package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"zukify.com/database"
)

type FlowData struct {
	Name  string          `json:"name"`
	WID   string          `json:"wid"`
	Nodes json.RawMessage `json:"nodes"`
	Edges json.RawMessage `json:"edges"`
}

func SaveFlow(c echo.Context) error {
	user := c.Get("user").(jwt.MapClaims)
	uid, ok := user["uid"].(float64)
	if !ok {
		log.Printf("Failed to extract UID from token: %v", user)
		return echo.NewHTTPError(http.StatusInternalServerError, "Invalid token")
	}

	var flowData FlowData

	if err := c.Bind(&flowData); err != nil {
		fmt.Println("Error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Convert flow data to JSON
	flowJSON, err := json.Marshal(map[string]json.RawMessage{
		"nodes": flowData.Nodes,
		"edges": flowData.Edges,
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to marshal flow data"})
	}

	// Construct the table name
	tableName := fmt.Sprintf("%s_flow", flowData.WID)
	fmt.Println("Table Name: ", tableName)
	query := fmt.Sprintf(`
		INSERT INTO %s (name, flow_data, modified_by)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
		name = VALUES(name),
		flow_data = VALUES(flow_data),
		modified_by = VALUES(modified_by)
	`, tableName)
	_, err = database.WorkspaceDB.Exec(query, flowData.Name, flowJSON, uid)

	if err != nil {
		fmt.Println("Error", err)

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save flow data: " + err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Flow data saved successfully"})
}

func LoadSpecificFlow(c echo.Context) error {
	// Get workspace ID and flow ID from the request
	// This assumes you're passing these as query parameters or path parameters
	workspaceID := c.QueryParam("wid") // or c.QueryParam("wid")
	flowID := c.QueryParam("fid")      // or c.QueryParam("fid")

	// Construct the table name using the workspace ID
	tableName := fmt.Sprintf("%s_flow", workspaceID)
	fmt.Println("Table Name:", tableName)
	// Retrieve flow data from the database
	query := fmt.Sprintf("SELECT flow_data, fid FROM %s WHERE fid = ?", tableName)
	var flowJSON []byte
	var fid string
	err := database.WorkspaceDB.QueryRow(query, flowID).Scan(&flowJSON, &fid)

	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Error", err)
			// If no data is found, return an appropriate error message
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Flow not found"})
		}
		fmt.Println("Error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load flow data"})
	}

	var flowData FlowData
	if err := json.Unmarshal(flowJSON, &flowData); err != nil {
		fmt.Println("Error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to unmarshal flow data"})
	}

	return c.JSON(http.StatusOK, flowData)
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