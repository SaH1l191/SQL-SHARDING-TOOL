package connections

import (
	"context"
	"database/sql" 
	"time"
)
 

func newConnectionSetup(ctx context.Context, connString string) (*sql.DB, error) {
	db, err := sql.Open("mysql", connString)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)

	//need to verify the sql connection as Open just verifies sql string and ping connects to the database
	pingConn, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := db.PingContext(pingConn); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
