package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"fmt"
	"github.com/labstack/echo/v4"
	"zukify.com/database"
)

type FlowData struct {
	Nodes []interface{} `json:"nodes"`
	Edges []interface{} `json:"edges"`
}

func SaveFlow(c echo.Context) error {
	var flowData FlowData

	if err := c.Bind(&flowData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Get user ID and workspace ID from the token
	// This is a placeholder - implement your own logic to get these IDs
	userID := 1
	workspaceID := 1

	// Convert flow data to JSON
	flowJSON, err := json.Marshal(flowData)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to marshal flow data"})
	}

	// Save flow data to the database
	query := `INSERT INTO flow_data (user_id, workspace_id, flow_json) VALUES (?, ?, ?)
              ON DUPLICATE KEY UPDATE flow_json = VALUES(flow_json)`
	
	_, err = database.WorkspaceDB.Exec(query, userID, workspaceID, flowJSON)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save flow data"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Flow data saved successfully"})
}

func LoadFlow(c echo.Context) error {
	// Get user ID and workspace ID from the token
	// This is a placeholder - implement your own logic to get these IDs
	userID := 1
	workspaceID := 1

	// Retrieve flow data from the database
	query := `SELECT flow_json FROM flow_data WHERE user_id = ? AND workspace_id = ?`
	var flowJSON []byte
	err := database.WorkspaceDB.QueryRow(query, userID, workspaceID).Scan(&flowJSON)

	if err != nil {
		if err == sql.ErrNoRows {
			// If no data is found, return an empty flow
			return c.JSON(http.StatusOK, FlowData{Nodes: []interface{}{}, Edges: []interface{}{}})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load flow data"})
	}

	var flowData FlowData
	if err := json.Unmarshal(flowJSON, &flowData); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to unmarshal flow data"})
	}

	return c.JSON(http.StatusOK, flowData)
}
func LoadSpecificFlow(c echo.Context) error {
	// Get workspace ID and flow ID from the request
	// This assumes you're passing these as query parameters or path parameters
	workspaceID := c.QueryParam("wid") // or c.QueryParam("wid")
	flowID := c.QueryParam("fid")      // or c.QueryParam("fid")

	// Construct the table name using the workspace ID
	tableName := fmt.Sprintf("%s_flow", workspaceID)
	fmt.Println("Table Name:",tableName)
	// Retrieve flow data from the database
	query := fmt.Sprintf("SELECT flow_data, fid FROM %s WHERE fid = ?", tableName)
	var flowJSON []byte
	var fid string
	err := database.WorkspaceDB.QueryRow(query, flowID).Scan(&flowJSON, &fid)

	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Error",err)
			// If no data is found, return an appropriate error message
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Flow not found"})
		}
		fmt.Println("Error",err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load flow data"})
	}

	var flowData FlowData
	if err := json.Unmarshal(flowJSON, &flowData); err != nil {
		fmt.Println("Error",err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to unmarshal flow data"})
	}
	

	return c.JSON(http.StatusOK, flowData)
}