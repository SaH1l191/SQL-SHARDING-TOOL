package db

import (
	"fmt"
	"sqlsharder/internal/config"
)

//will return the postgres connection string
func BuildDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.AppDBCreds.DB_USER,
		config.AppDBCreds.DB_PASSWORD,
		config.AppDBCreds.DB_HOST,
		config.AppDBCreds.DB_PORT,
		config.AppDBCreds.DB_NAME)
}

//eg  : postgres://user:password@host:port/dbname?sslmode=disable
