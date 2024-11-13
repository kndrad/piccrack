package cmd

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

func initLogger() {
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	} else {
		slog.SetLogLoggerLevel(slog.LevelError)
	}
}
