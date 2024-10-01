package database

import (
	"fmt"
)

func SaveATResponse(wid, id string, response string) error {
	query := fmt.Sprintf("UPDATE %s_at SET response = ? WHERE id = ?", wid)
	_, err := WorkspaceDB.Exec(query, response, id)
	return err
}