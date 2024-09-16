package database

import (
	"fmt"
	"log"
	"database/sql"
)

func CreateDataTable(tablePrefix string, uid int) error {
	_, err := WorkspaceDB.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s_data (
			uid INT(11) NULL,
			role INT(11) NULL
		)
	`, tablePrefix))
	if err != nil {
		log.Printf("Failed to create Data table: %v", err)
		return err
	}

	// Insert user into Data table
	_, err = WorkspaceDB.Exec(fmt.Sprintf("INSERT INTO %s_data (uid, role) VALUES (?, ?)", tablePrefix), uid, 1)
	if err != nil {
		log.Printf("Failed to insert user into Data table: %v", err)
		return err
	}

	return nil
}

func CheckUserRole(uid int, wid string) (bool, error) {
	var role int
	err := WorkspaceDB.QueryRow(fmt.Sprintf("SELECT role FROM %s_data WHERE uid = ?", wid), uid).Scan(&role)
	if err != nil {
		return false, err
	}
	return role == 1, nil
}

func AddCollaborator(wid string, collabUID int) error {
	// First, check if the collaborator already exists in the workspace
	var existingUID int
	err := WorkspaceDB.QueryRow(fmt.Sprintf("SELECT uid FROM %s_data WHERE uid = ?", wid), collabUID).Scan(&existingUID)
	
	if err == nil {
		// Collaborator already exists
		return fmt.Errorf("collaborator already exists in this workspace")
	} else if err != sql.ErrNoRows {
		// An unexpected error occurred
		log.Printf("Error checking for existing collaborator: %v", err)
		return err
	}

	// If we reach here, the collaborator doesn't exist, so we can add them
	_, err = WorkspaceDB.Exec(fmt.Sprintf("INSERT INTO %s_data (uid, role) VALUES (?, ?)", wid), collabUID, 0)
	if err != nil {
		log.Printf("Error adding collaborator: %v", err)
		return err
	}

	return nil
}