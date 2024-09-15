package database

import (
	"database/sql"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UID      int             `json:"uid"`
	Username string          `json:"username"`
	Password string          `json:"password,omitempty"` // Changed to omitempty
	Devices  json.RawMessage `json:"devices,omitempty"`
}

func CreateUser(user *User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Convert devices to JSON string
	devicesJSON, err := json.Marshal(user.Devices)
	if err != nil {
		return err
	}

	_, err = UserDB.Exec("INSERT INTO auth (username, password, devices) VALUES (?, ?, ?)",
		user.Username, string(hashedPassword), string(devicesJSON))
	return err
}

func UserExists(username string) (bool, error) {
	var exists bool
	err := UserDB.QueryRow("SELECT EXISTS(SELECT 1 FROM auth WHERE username = ?)", username).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}


func GetUserByUsername(username string) (*User, error) {
	user := &User{}
	var devices sql.NullString
	err := UserDB.QueryRow("SELECT uid, username, password, devices FROM auth WHERE username = ?", username).
		Scan(&user.UID, &user.Username, &user.Password, &devices)
	if err != nil {
		return nil, err
	}
	if devices.Valid {
		user.Devices = json.RawMessage(devices.String)
	}
	return user, nil
}