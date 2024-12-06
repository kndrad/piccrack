package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type key string

const (
	durationKey = key("duration")
)

func LogTime(next http.HandlerFunc, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.InfoContext(r.Context(), "Received request", slog.String("url", r.URL.String()))

		start := time.Now()
		next(w, r)
		elapsed := time.Since(start)

		logger.Info("Finished", slog.Duration("duration", elapsed))
	}
}
