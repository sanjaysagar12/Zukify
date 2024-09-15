package database

import (
	"fmt"
	"log"
)

func CreateATTable(tablePrefix string) error {
	_, err := WorkspaceDB.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s_at (
			id INT(11) AUTO_INCREMENT PRIMARY KEY,
			path TINYTEXT NULL,
			tag TINYTEXT NULL,
			Method VARCHAR(6) NULL,
			url TINYTEXT NULL,
			header LONGTEXT NULL,
			body LONGTEXT NULL,
			testcases LONGTEXT NULL,
			response TEXT NULL,
			modified_by INT(11) NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)
	`, tablePrefix))
	if err != nil {
		log.Printf("Failed to create AT table: %v", err)
		return err
	}
	return nil
}