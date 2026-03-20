package config

import "database/sql"

// go get github.com/lib/pq v1.11.2 // indirect
// go get github.com/pganalyze/pg_query_go/v5 v5.1.0

type ApplicationDBCreds struct {
	DB_HOST     string
	DB_PASSWORD string
	DB_NAME     string
	DB_USER     string
	DB_PORT     string
}

var AppDBCreds ApplicationDBCreds

type ApplicationDatabaseConn struct {
	ConnInstance *sql.DB
}

var ApplicationDatabaseConnection ApplicationDatabaseConn

//will be instantiated in main.go
