package cmd

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func initLogger() {
	Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	} else {
		slog.SetLogLoggerLevel(slog.LevelError)
	}
}
