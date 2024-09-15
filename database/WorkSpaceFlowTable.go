package database

import (
	"fmt"
	"database/sql"
	"log"
)

type FlowData struct {
	Name     string `json:"name"`
	FlowData string `json:"flow_data"`
	NodeData string `json:"node_data"`
}

type PathFlowData struct {
	FID  int    `json:"fid"`
	Name string `json:"name"`
}

type AllFlowData struct {
	FID      int    `json:"fid"`
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


func FetchPathFlow(wid string) ([]PathFlowData, error) {
	query := fmt.Sprintf("SELECT fid, name FROM %s_flow", wid)
	rows, err := WorkspaceDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PathFlowData
	for rows.Next() {
		var data PathFlowData
		if err := rows.Scan(&data.FID, &data.Name); err != nil {
			return nil, err
		}
		result = append(result, data)
	}

	return result, nil
}

func FetchAllFlow(wid, fid string) (*AllFlowData, error) {
	query := fmt.Sprintf("SELECT fid, name, flow_data, node_data FROM %s_flow WHERE fid = ?", wid)
	var data AllFlowData
	err := WorkspaceDB.QueryRow(query, fid).Scan(
		&data.FID, &data.Name, &data.FlowData, &data.NodeData,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &data, nil
}