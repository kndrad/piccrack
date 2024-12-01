package cmd

import (
	"log/slog"
	"os"
)

func DefaultLogger(verbose bool) *slog.Logger {
	l := slog.New(slog.NewTextHandler(os.Stdout, nil))

	slog.SetLogLoggerLevel(slog.LevelError)
	if verbose {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	return l
}
