package db

import (
	"database/sql"
	"fmt"
	"sqlsharder/internal/config"
)

//func for initial startup to check if DB exists, if not create it to prevent throwing
// errors when trying to connect to it
func EnsureDatabaseExists() error {
	sysConnStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/postgres?sslmode=disable",
		config.AppDBCreds.DB_USER,
		config.AppDBCreds.DB_PASSWORD,
		config.AppDBCreds.DB_HOST,
		config.AppDBCreds.DB_PORT,
	)
	sysDB, err := sql.Open("postgres", sysConnStr)

	if err != nil {
		return fmt.Errorf("failed to connect to system database: %w", err)
	}

	var exists bool
	err = sysDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", config.AppDBCreds.DB_NAME).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if !exists {
		query := fmt.Sprintf("CREATE DATABASE %s", config.AppDBCreds.DB_NAME)
		_, err = sysDB.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
	}
	defer sysDB.Close()
	return nil
}
