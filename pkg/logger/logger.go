package logger

import (
	"log/slog"
	"os"
)

//centralized logger for the application
var Logger *slog.Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
