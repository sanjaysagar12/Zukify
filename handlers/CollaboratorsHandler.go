package handlers

import (
	"net/http"
	"log"
	// "fmt"

	"github.com/labstack/echo/v4"
	"github.com/golang-jwt/jwt"
	"zukify.com/database"
)

func HandlerAddCollaborator(c echo.Context) error {
	user := c.Get("user").(jwt.MapClaims)
	uid, ok := user["uid"].(float64)
	if !ok {
		log.Printf("Failed to extract UID from token: %v", user)
		return echo.NewHTTPError(http.StatusInternalServerError, "Invalid token")
	}

	var req struct {
		WID       string `json:"wid"`
		CollabUID int    `json:"collab_uid"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Check if the user has admin role (role 1) in the workspace
	hasAdminRole, err := database.CheckUserRole(int(uid), req.WID)
	if err != nil {
		log.Printf("Failed to check user role: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify user role")
	}
	if !hasAdminRole {
		return echo.NewHTTPError(http.StatusForbidden, "You don't have permission to add collaborators")
	}

	// Add collaborator to the workspace
	err = database.AddCollaborator(req.WID, req.CollabUID)
	if err != nil {
		if err.Error() == "collaborator already exists in this workspace" {
			return echo.NewHTTPError(http.StatusConflict, "Collaborator already exists in this workspace")
		}
		log.Printf("Failed to add collaborator: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to add collaborator")
	}

	// Get the full workspace info
	workspaceInfo, err := database.GetWorkspaceInfo(int(uid), req.WID)
	if err != nil {
		log.Printf("Failed to get workspace info: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get workspace info")
	}

	// Add workspace to collaborator's user record
	err = database.AddWorkspaceToCollaborator(req.CollabUID, workspaceInfo)
	if err != nil {
		log.Printf("Failed to add workspace to collaborator: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update collaborator's workspace list")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Collaborator added successfully",
	})
}