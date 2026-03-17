package loader

import ( 
	"database/sql"
	"sqlsharder/internal/db"
	"sqlsharder/pkg/logger"
	"sqlsharder/pkg/environment"
)

// sequencing file
//loading service in order


func Loadservices() error { 
	err := environment.LoadEnv()
	if err != nil {
		logger.Logger.Error("Failed to load environment","error", err)
		return err
	}
	environment.LoadEnvVariables()

	err = db.EnsureDatabaseExists()
	if err != nil {
		//message , key , value
		logger.Logger.Error("Failed to ensure database exists", "error", err)
		return err
	}
	dsn := db.BuildDSN()
	err = db.ApplyMigrations(dsn)
	if err != nil {
		logger.Logger.Error("Failed to apply migrations", "error", err)
		return err
	}
	


	logger.Logger.Info("Env Loaded & Database Created", "status", "success")

	return nil
}


func LoadApplicationDatabase() (*sql.DB, error) {
	db, err := db.LoadApplicationDatabase()
	if err != nil {
		return nil, err
	}
	return db, nil
}