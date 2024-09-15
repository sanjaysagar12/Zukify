package database

import (
	"fmt"
	"log"
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