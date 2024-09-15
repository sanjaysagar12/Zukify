package database

import (
	"fmt"
	"log"
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