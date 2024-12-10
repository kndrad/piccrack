package logger

import (
	"log/slog"
	"os"
)

func New(verbose bool) *slog.Logger {
	l := slog.New(slog.NewTextHandler(os.Stdout, nil))

	slog.SetLogLoggerLevel(slog.LevelError)
	if verbose {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	return l
}
