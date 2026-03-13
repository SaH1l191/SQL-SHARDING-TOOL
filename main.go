package main

import (
	"context"
	"sqlsharder/pkg/logger"
)

func main() { 
	ctx,cancel := context.WithCancel(context.Background())
	defer cancel()
	
	app := NewApp() 
	app.Run(ctx)

	logger.Logger.Info("Application started") 
 
}
