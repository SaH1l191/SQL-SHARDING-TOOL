// load all migrations
package db

import (
	"fmt"
	"os"
	"path/filepath"
	"sqlsharder/pkg/logger"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func ApplyMigrations(dsn string) error {

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	migrationsPath := filepath.Join(wd, "migrations")

	sourceURL := "file://" + filepath.ToSlash(migrationsPath)

	fmt.Println("Working directory:", wd)
	fmt.Println("Migrations path:", migrationsPath)
	fmt.Println("Source URL:", sourceURL)

	m, err := migrate.New(
		sourceURL,
		dsn,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	logger.Logger.Info("Migrations applied successfully", "status", "success")
	return nil
}
