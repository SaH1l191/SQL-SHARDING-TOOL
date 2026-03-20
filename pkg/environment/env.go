package environment

import (
	"os"
	"sqlsharder/internal/config"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}
	return nil
}

func LoadEnvVariables() {
	config.AppDBCreds.DB_HOST = os.Getenv("DB_HOST")
	config.AppDBCreds.DB_PASSWORD = os.Getenv("DB_PASSWORD")
	config.AppDBCreds.DB_NAME = os.Getenv("DB_NAME")
	config.AppDBCreds.DB_USER = os.Getenv("DB_USER")
	config.AppDBCreds.DB_PORT = os.Getenv("DB_PORT")
}
