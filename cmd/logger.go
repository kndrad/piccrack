package cmd

import (
	"log/slog"
	"os"
)

var DefaultLogger *slog.Logger

func initLogger() {
	DefaultLogger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	} else {
		slog.SetLogLoggerLevel(slog.LevelError)
	}
}
