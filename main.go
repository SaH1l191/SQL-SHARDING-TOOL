package main

import (
	"context"
	"os/signal"
	"sqlsharder/pkg/logger"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	//listens for OS signals:
	// SIGINT → usually sent when you press Ctrl+C
	// SIGTERM → sent when the system asks your app to terminate (e.g., during shutdown, container stop)
	// When any of those signals arrive, the returned ctx is automatically cancelled
	defer stop()

	app := NewApp()
	if err := app.Run(ctx); err != nil {
		logger.Logger.Error("Application error", "error", err)
	}
}
