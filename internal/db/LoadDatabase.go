package db
import (
	"database/sql"
	"fmt"
	"sqlsharder/internal/config"
	_ "github.com/lib/pq"
	// blank import.
)

func LoadApplicationDatabase() (*sql.DB, error) {
	connUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
	config.AppDBCreds.DB_USER, 
	config.AppDBCreds.DB_PASSWORD, 
	config.AppDBCreds.DB_HOST, 
	config.AppDBCreds.DB_PORT, 
	config.AppDBCreds.DB_NAME)

	conn , err := sql.Open("postgres", connUrl)
	if err != nil {
		return nil ,err 
	}
	err = conn.Ping()
	if err != nil {
		return nil ,err 
	}

	return conn, nil
}