package main

import (
	"context"
	"database/sql" 
	"sqlsharder/internal/loader"
	"sqlsharder/pkg/logger"
)

type App struct {
	ctx context.Context
	db  *sql.DB
}

 
func NewApp() *App {
	return &App{
		ctx: context.Background(),
	}
}

 
func (a *App) Run(ctx context.Context) { 
	a.ctx = ctx
	err := loader.Loadservices()
	if err != nil {
		logger.Logger.Error("Failed to load core services", "error", err)
		panic(err)
	}
 
	a.db, err = loader.LoadApplicationDatabase()
	if err != nil {
		logger.Logger.Error("Failed to load application database", "error", err)
		panic(err)
	}
	 
}
  

