package main
 
import (
	"context"
	"os/signal"
	"sqlsharder/pkg/logger"
	"syscall"
)
 
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
 
	app := NewApp()
	if err := app.Run(ctx); err != nil { 
		logger.Logger.Error("Application error", "error", err)
	}
}
 