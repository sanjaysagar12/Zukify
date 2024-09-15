package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/go-sql-driver/mysql"
)

var (
	UserDB      *sql.DB
	WorkspaceDB *sql.DB
)

func InitDB() error {
	// Load the .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file")
		return err
	}

	var err error
	UserDB, err = ConnectUserDB()
	if err != nil {
		return fmt.Errorf("failed to connect to User database: %v", err)
	}

	WorkspaceDB, err = ConnectWorkspaceDB()
	if err != nil {
		return fmt.Errorf("failed to connect to Workspace database: %v", err)
	}

	return nil
}

// ConnectUserDB establishes a connection to the User database
func ConnectUserDB() (*sql.DB, error) {
	return connectToDB("USER")
}

// ConnectWorkspaceDB establishes a connection to the Workspace database
func ConnectWorkspaceDB() (*sql.DB, error) {
	return connectToDB("WORKSPACE")
}

// connectToDB is a helper function to connect to a specific database
func connectToDB(dbType string) (*sql.DB, error) {
	dbHost := os.Getenv("MYSQL_HOST")
	dbPort := os.Getenv("MYSQL_PORT")
	dbUser := os.Getenv("MYSQL_USER")
	dbPass := os.Getenv("MYSQL_PASSWORD")
	dbName := os.Getenv(fmt.Sprintf("MYSQL_%s_DBNAME", dbType))

	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s database: %v", dbType, err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping %s database: %v", dbType, err)
	}

	fmt.Printf("Successfully connected to %s database\n", dbType)
	return db, nil
}