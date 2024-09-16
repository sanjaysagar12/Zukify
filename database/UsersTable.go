package database

import (
	"encoding/json"
	"log"
	"database/sql"
	"errors"
	"fmt"
)

type WorkspaceInfo struct {
	WID  string `json:"wid"`
	Name string `json:"name"`
}

var ErrWorkspaceExists = errors.New("workspace with this name already exists")

func AddWorkspaceToUser(uid int, hashedValue, workspaceName string) error {
	var workspaces sql.NullString
	err := UserDB.QueryRow("SELECT workspace FROM users WHERE uid = ?", uid).Scan(&workspaces)
	if err != nil {
		if err == sql.ErrNoRows {
			return createNewUserWithWorkspace(uid, hashedValue, workspaceName)
		}
		log.Printf("Failed to get current workspaces: %v", err)
		return err
	}

	var workspaceList []WorkspaceInfo
	if workspaces.Valid && workspaces.String != "" {
		err = json.Unmarshal([]byte(workspaces.String), &workspaceList)
		if err != nil {
			log.Printf("Failed to unmarshal workspaces: %v", err)
			return err
		}

		// Check if workspace with the same name already exists
		for _, ws := range workspaceList {
			if ws.Name == workspaceName {
				return ErrWorkspaceExists
			}
		}
	}

	// Add new workspace
	workspaceList = append(workspaceList, WorkspaceInfo{WID: hashedValue, Name: workspaceName})

	// Marshal updated workspace list
	updatedWorkspaces, err := json.Marshal(workspaceList)
	if err != nil {
		log.Printf("Failed to marshal updated workspaces: %v", err)
		return err
	}

	// Update users table
	_, err = UserDB.Exec("UPDATE users SET workspace = ? WHERE uid = ?", string(updatedWorkspaces), uid)
	if err != nil {
		log.Printf("Failed to update user's workspace: %v", err)
		return err
	}

	return nil
}

func createNewUserWithWorkspace(uid int, hashedValue, workspaceName string) error {
	var username string
	err := UserDB.QueryRow("SELECT username FROM auth WHERE uid = ?", uid).Scan(&username)
	if err != nil {
		log.Printf("Failed to get username from auth table: %v", err)
		return err
	}

	workspaceList := []WorkspaceInfo{{WID: hashedValue, Name: workspaceName}}
	workspaces, err := json.Marshal(workspaceList)
	if err != nil {
		log.Printf("Failed to marshal new workspace: %v", err)
		return err
	}

	_, err = UserDB.Exec("INSERT INTO users (uid, username, workspace) VALUES (?, ?, ?)", uid, username, string(workspaces))
	if err != nil {
		log.Printf("Failed to create new user with workspace: %v", err)
		return err
	}

	return nil
}

func WorkspaceExists(uid int, workspaceName string) (bool, error) {
	var workspaces sql.NullString
	err := UserDB.QueryRow("SELECT workspace FROM users WHERE uid = ?", uid).Scan(&workspaces)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		log.Printf("Failed to get current workspaces: %v", err)
		return false, err
	}

	if workspaces.Valid && workspaces.String != "" {
		var workspaceList []WorkspaceInfo
		err = json.Unmarshal([]byte(workspaces.String), &workspaceList)
		if err != nil {
			log.Printf("Failed to unmarshal workspaces: %v", err)
			return false, err
		}

		for _, ws := range workspaceList {
			if ws.Name == workspaceName {
				return true, nil
			}
		}
	}

	return false, nil
}

func GetUserWorkspaces(uid int) ([]WorkspaceInfo, error) {
	var workspaces sql.NullString
	err := UserDB.QueryRow("SELECT workspace FROM users WHERE uid = ?", uid).Scan(&workspaces)
	if err != nil {
		return nil, err
	}

	if !workspaces.Valid || workspaces.String == "" {
		return nil, nil
	}

	var workspaceList []WorkspaceInfo
	err = json.Unmarshal([]byte(workspaces.String), &workspaceList)
	if err != nil {
		return nil, err
	}

	return workspaceList, nil
}

func AddWorkspaceToCollaborator(collabUID int, workspace WorkspaceInfo) error {
	var workspaces sql.NullString
	err := UserDB.QueryRow("SELECT workspace FROM users WHERE uid = ?", collabUID).Scan(&workspaces)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create new user entry if it doesn't exist
			return createNewUserWithWorkspace(collabUID, workspace.WID, workspace.Name)
		}
		return err
	}

	var workspaceList []WorkspaceInfo
	if workspaces.Valid && workspaces.String != "" {
		err = json.Unmarshal([]byte(workspaces.String), &workspaceList)
		if err != nil {
			return err
		}
	}

	// Add new workspace
	workspaceList = append(workspaceList, workspace)

	// Marshal updated workspace list
	updatedWorkspaces, err := json.Marshal(workspaceList)
	if err != nil {
		return err
	}

	// Update users table
	_, err = UserDB.Exec("UPDATE users SET workspace = ? WHERE uid = ?", string(updatedWorkspaces), collabUID)
	return err
}

func GetWorkspaceInfo(uid int, wid string) (WorkspaceInfo, error) {
	workspaces, err := GetUserWorkspaces(uid)
	if err != nil {
		return WorkspaceInfo{}, err
	}

	for _, ws := range workspaces {
		if ws.WID == wid {
			return ws, nil
		}
	}

	return WorkspaceInfo{}, fmt.Errorf("workspace not found")
}

