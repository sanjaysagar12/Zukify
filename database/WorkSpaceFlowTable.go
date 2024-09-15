package database

import (
	"fmt"
	"log"
)

func CreateFlowTable(tablePrefix string) error {
	_, err := WorkspaceDB.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s_flow (
			fid INT(11) AUTO_INCREMENT PRIMARY KEY,
			name TINYTEXT NULL,
			flow_data LONGTEXT NULL,
			node_data LONGTEXT NULL,
			modified_by INT(11) NULL
		)
	`, tablePrefix))
	if err != nil {
		log.Printf("Failed to create Flow table: %v", err)
		return err
	}
	return nil
}