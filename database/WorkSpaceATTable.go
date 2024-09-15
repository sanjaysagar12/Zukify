package database

import (
	"fmt"
	"log"
	"database/sql"
	"encoding/json"
)

type ATData struct {
	Path      string `json:"path"`
	Tag       string `json:"tag"`
	Method    string `json:"method"`
	URL       string `json:"url"`
	Header    string `json:"header"`
	Body      string `json:"body"`
	Testcases string `json:"testcases"`
	Response  string `json:"response"`
}

type PathATData struct {
	ID     int    `json:"id"`
	Path   string `json:"path"`
	Method string `json:"method"`
}

type AllATData struct {
	ID        int    `json:"id"`
	Path      string `json:"path"`
	Tag       string `json:"tag"`
	Method    string `json:"method"`
	URL       string `json:"url"`
	Header    string `json:"header"`
	Body      string `json:"body"`
	Testcases string `json:"testcases"`
	Response  string `json:"response"`
}

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

func SaveATData(tablePrefix string, data *ATData, uid int) error {
	_, err := WorkspaceDB.Exec(fmt.Sprintf(`
		INSERT INTO %s_at (path, tag, Method, url, header, body, testcases, response, modified_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, tablePrefix), data.Path, data.Tag, data.Method, data.URL, data.Header, data.Body, data.Testcases, data.Response, uid)

	return err
}

func FetchPathAT(wid string) ([]PathATData, error) {
	query := fmt.Sprintf("SELECT id, path, Method FROM %s_at", wid)
	rows, err := WorkspaceDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PathATData
	for rows.Next() {
		var data PathATData
		if err := rows.Scan(&data.ID, &data.Path, &data.Method); err != nil {
			return nil, err
		}
		result = append(result, data)
	}

	return result, nil
}

func FetchAllAT(wid, id string) (*AllATData, error) {
	query := fmt.Sprintf("SELECT id, path, tag, Method, url, header, body, testcases, response FROM %s_at WHERE id = ?", wid)
	var data AllATData
	err := WorkspaceDB.QueryRow(query, id).Scan(
		&data.ID, &data.Path, &data.Tag, &data.Method, &data.URL,
		&data.Header, &data.Body, &data.Testcases, &data.Response,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func UserHasAccessToWorkspace(uid int, wid string) (bool, error) {
	var workspaces string
	err := UserDB.QueryRow("SELECT workspace FROM users WHERE uid = ?", uid).Scan(&workspaces)
	if err != nil {
		return false, err
	}

	var workspaceList []struct {
		WID  string `json:"wid"`
		Name string `json:"name"`
	}
	err = json.Unmarshal([]byte(workspaces), &workspaceList)
	if err != nil {
		return false, err
	}

	for _, ws := range workspaceList {
		if ws.WID == wid {
			return true, nil
		}
	}

	return false, nil
}