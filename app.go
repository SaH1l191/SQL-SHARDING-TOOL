package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"sqlsharder/internal/config"
	"sqlsharder/internal/connections"
	"sqlsharder/internal/handler"
	"sqlsharder/internal/loader"
	"sqlsharder/internal/repository"
	"sqlsharder/internal/router"
	"sqlsharder/internal/service"
	"sqlsharder/pkg/logger"
)

type App struct {
	db          *sql.DB
	server      *http.Server
	connStore   *connections.ConnectionStore
	connManager *connections.ConnectionManager
}

func NewApp() *App {
	return &App{}
}

func (a *App) Run(ctx context.Context) error {
	if err := loader.Loadservices(); err != nil {
		logger.Logger.Error("Failed to load core services", "error", err)
		return err
	}

	var err error
	a.db, err = loader.LoadApplicationDatabase()
	if err != nil {
		logger.Logger.Error("Failed to load application database", "error", err)
		return err
	}

	config.ApplicationDatabaseConnection.ConnInstance = a.db

	logger.Logger.Info("Database connection established", "host", config.AppDBCreds.DB_HOST, "database", config.AppDBCreds.DB_NAME, "port", config.AppDBCreds.DB_PORT)
	a.server = a.buildServer()

	logger.Logger.Info("Starting HTTP server", "addr", ":8080")
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Error("HTTP server error", "error", err)
		}
	}()

	<-ctx.Done()
	//pauses here indefinitely unless the signal is received or catched by NotfiyContext in main.go
	logger.Logger.Info("Shutdown signal received, cleaning up...")
	return a.cleanup()
}

// buildServer wires repository → service → handler → router.
// Only constructor calls live here — no logic, no conditionals.
func (a *App) buildServer() *http.Server {
	// handler -> service -> repository
	projectRepo := repository.NewProjectRepository(a.db)
	projectService := service.NewProjectService(projectRepo)
	projectHandler := handler.NewProjectHandler(projectService)

	shardRepo := repository.NewShardRepository(a.db)
	shardService := service.NewShardService(shardRepo)
	shardHandler := handler.NewShardHandler(shardService)

	schemaRepo := repository.NewProjectSchemaRepository(a.db)
	schemaService := service.NewProjectSchemaService(schemaRepo)
	schemaHandler := handler.NewProjectSchemaHandler(schemaService)

	shardConnRepo := repository.NewShardConnectionRepository(a.db)
	shardConnService := service.NewShardConnectionService(shardConnRepo)
	shardConnHandler := handler.NewShardConnectionHandler(shardConnService)

	a.connStore = connections.NewConnectionStore()
	a.connManager = connections.NewConnectionManager(
		a.connStore,
		projectRepo,
		shardRepo,
		shardConnRepo,
	)

	r := router.New()
	r.RegisterHealthRoute()
	r.RegisterProjectRoutes(projectHandler)
	r.RegisterShardRoutes(shardHandler)
	r.RegisterProjectSchemaRoutes(schemaHandler)
	r.RegisterShardConnectionRoutes(shardConnHandler)

	return &http.Server{
		Addr:    ":8080",
		Handler: r.Engine(),
	}
}

func (a *App) cleanup() error {
	if a.server != nil {
		logger.Logger.Info("Shutting down HTTP server...")
		if err := a.server.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("server shutdown: %w", err)
		}
		logger.Logger.Info("HTTP server stopped")
	}

	if a.db != nil {
		if err := a.db.Close(); err != nil {
			return fmt.Errorf("close database: %w", err)
		}
		logger.Logger.Info("Database connection closed")
	}

	logger.Logger.Info("Shutdown complete")
	return nil
}
