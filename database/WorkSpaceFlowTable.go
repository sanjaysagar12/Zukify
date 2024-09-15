package database

import (
	"fmt"
	"log"
)

type FlowData struct {
	Name     string `json:"name"`
	FlowData string `json:"flow_data"`
	NodeData string `json:"node_data"`
}

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

func SaveFlowData(tablePrefix string, data *FlowData, uid int) error {
	_, err := WorkspaceDB.Exec(fmt.Sprintf(`
		INSERT INTO %s_flow (name, flow_data, node_data, modified_by)
		VALUES (?, ?, ?, ?)
	`, tablePrefix), data.Name, data.FlowData, data.NodeData, uid)

	return err
}